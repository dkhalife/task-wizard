package caldav

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"dkhalife.com/tasks/core/internal/models"
	tRepo "dkhalife.com/tasks/core/internal/repos/task"
	uRepo "dkhalife.com/tasks/core/internal/repos/user"
	"dkhalife.com/tasks/core/internal/services/logging"
)

type CalDavRepository struct {
	tRepo *tRepo.TaskRepository
	uRepo uRepo.IUserRepo
}

func NewCalDavRepository(tRepo *tRepo.TaskRepository, uRepo uRepo.IUserRepo) *CalDavRepository {
	return &CalDavRepository{tRepo: tRepo, uRepo: uRepo}
}

func formatHTTPDate(t time.Time) string {
	return t.UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
}

func generateETag(task *models.Task) string {
	// Calculate the latest modification timestamp, this will
	// cover all cases where data might have changed in the task
	// or its associated labels.
	latest := task.UpdatedAt
	if latest == nil {
		latest = &task.CreatedAt
	}

	labelIDs := make([]string, 0, len(task.Labels))
	for _, label := range task.Labels {
		labelIDs = append(labelIDs, strconv.Itoa(label.ID))
		if label.UpdatedAt != nil && label.UpdatedAt.After(*latest) {
			latest = label.UpdatedAt
		} else if label.CreatedAt.After(*latest) {
			latest = &label.CreatedAt
		}
	}

	base := latest.UTC().Format("20060102T150405Z")

	// Including the label ids covers the cases where association
	// of labels to tasks changes without requiring a new date
	// field update.
	if len(labelIDs) > 0 {
		sort.Strings(labelIDs) // Ensure labelIDs are sorted for consistent ETag generation
		base = base + ";" + strings.Join(labelIDs, ";")
	}

	hash := md5.Sum([]byte(base))
	return fmt.Sprintf("\"%s\"", hex.EncodeToString(hash[:]))
}

func generateCategories(task *models.Task) string {
	if len(task.Labels) == 0 {
		return ""
	}

	labels := make([]string, len(task.Labels))
	for i, label := range task.Labels {
		labels[i] = label.Name
	}
	return strings.Join(labels, ",")
}

func generateVTODO(task *models.Task) string {
	created := task.CreatedAt.UTC().Format("20060102T150405Z")

	var lastModified string
	if task.UpdatedAt != nil {
		lastModified = task.UpdatedAt.UTC().Format("20060102T150405Z")
	} else {
		lastModified = created
	}

	dueDate := ""
	if task.NextDueDate != nil {
		dueDate = fmt.Sprintf("DUE:%s", task.NextDueDate.UTC().Format("20060102T150405Z"))
	}

	categories := generateCategories(task)
	if len(categories) > 0 {
		categories = fmt.Sprintf("CATEGORIES:%s", categories)
	}

	vtodo := fmt.Sprintf(`BEGIN:VCALENDAR
PRODID:-//Mozilla.org/NONSGML Mozilla Calendar V1.1//EN
VERSION:2.0
BEGIN:VTODO
CREATED:%s
LAST-MODIFIED:%s
DTSTAMP:%s
UID:%d
SUMMARY:%s
%s
%s
PERCENT-COMPLETE:0
X-MOZ-GENERATION:1
END:VTODO
END:VCALENDAR`,
		created,
		lastModified,
		lastModified,
		task.ID,
		task.Title,
		dueDate,
		categories)

	return vtodo
}

func (r *CalDavRepository) PropfindTask(c context.Context, taskID int) (models.Multistatus, int, error) {
	task, err := r.tRepo.GetTask(c, taskID)
	if err != nil {
		return models.Multistatus{}, -1, err
	}

	etag := generateETag(task)
	vtodoContent := generateVTODO(task)
	contentLength := len(vtodoContent)

	response := models.Multistatus{
		DAVAttr:            models.DavNamespace,
		CalDAVAttr:         models.CalDavNamespace,
		CalendarServerAttr: models.CalendarServerNamespace,
		SabreAttr:          models.SabreNamespace,
		AppleAttr:          models.AppleNamespace,
		Responses: []models.Response{
			{
				Propstat: []models.Propstat{
					{
						Prop: models.Prop{
							GetLastModified:  formatHTTPDate(time.Now()),
							GetContentLength: contentLength,
							ResourceType:     &models.ResourceType{},
							GetETag:          etag,
							GetContentType:   "text/calendar; charset=utf-8; component=vtodo",
						},
						Status: "HTTP/1.1 200 OK",
					},
				},
			},
		},
	}

	return response, task.CreatedBy, nil
}

func (r *CalDavRepository) PropfindUserTasks(c context.Context, userID int) (models.Multistatus, error) {
	log := logging.FromContext(c)

	tasks, err := r.tRepo.GetTasks(c, userID)
	if err != nil {
		return models.Multistatus{}, err
	}

	response := models.Multistatus{
		DAVAttr:            models.DavNamespace,
		CalDAVAttr:         models.CalDavNamespace,
		CalendarServerAttr: models.CalendarServerNamespace,
		SabreAttr:          models.SabreNamespace,
		AppleAttr:          models.AppleNamespace,
		Responses:          []models.Response{},
	}

	lastModified, err := r.uRepo.GetLastCreatedOrModifiedForUserResources(c, userID)
	log.Debugf("Last modified for user %d: %s", userID, lastModified)
	if err != nil {
		return models.Multistatus{}, err
	}

	response.Responses = append(response.Responses, models.Response{
		Href: "/dav/tasks/",
		Propstat: []models.Propstat{
			{
				Prop: models.Prop{
					ResourceType: &models.ResourceType{
						Collection:  &struct{}{},
						Calendar:    &struct{}{},
						SharedOwner: &struct{}{},
					},
					GetCTag:   "http://sabre.io/ns/sync/" + lastModified,
					SyncToken: lastModified,
					SupportedComponents: &models.SupportedComponents{
						Comp: []models.CalComp{
							{Name: "VTODO"},
						},
					},
					DisplayName:         "Task Wizard",
					CalendarTimeZone:    "UTC",
					CalendarDescription: "",
					CalendarOrder:       "0",
					CalendarColor:       "",
				},
				Status: "HTTP/1.1 200 OK",
			},
		},
	})

	for _, task := range tasks {
		taskID := strconv.Itoa(task.ID)
		filename := taskID + ".ics"

		log.Debugf("Processing task for CalDAV: ID=%d, Title=%s", task.ID, task.Title)

		etag := generateETag(task)
		vtodoContent := generateVTODO(task)
		contentLength := len(vtodoContent)

		response.Responses = append(response.Responses, models.Response{
			Href: "/dav/tasks/" + filename,
			Propstat: []models.Propstat{
				{
					Prop: models.Prop{
						GetLastModified:  formatHTTPDate(time.Now()),
						GetContentLength: contentLength,
						ResourceType:     &models.ResourceType{},
						GetETag:          etag,
						GetContentType:   "text/calendar; charset=utf-8; component=vtodo",
					},
					Status: "HTTP/1.1 200 OK",
				},
			},
		})
	}

	return response, nil
}

func (r *CalDavRepository) GetTask(c context.Context, taskID int) (string, int, error) {
	task, err := r.tRepo.GetTask(c, taskID)
	if err != nil {
		return "", -1, err
	}

	return generateVTODO(task), task.CreatedBy, nil
}

func (r *CalDavRepository) MultiGet(c context.Context, request models.CalendarMultiget) (models.Multistatus, error) {
	response := models.Multistatus{
		DAVAttr:            models.DavNamespace,
		CalDAVAttr:         models.CalDavNamespace,
		CalendarServerAttr: models.CalendarServerNamespace,
		SabreAttr:          models.SabreNamespace,
		AppleAttr:          models.AppleNamespace,
		Responses:          []models.Response{},
	}

	for _, href := range request.Hrefs {
		filename := path.Base(href)
		taskID, err := strconv.Atoi(strings.TrimSuffix(filename, ".ics"))
		if err != nil {
			return models.Multistatus{}, fmt.Errorf("invalid task ID: %s", err.Error())
		}

		task, err := r.tRepo.GetTask(c, taskID)
		if err != nil {
			return models.Multistatus{}, err
		}

		etag := generateETag(task)
		vtodoContent := generateVTODO(task)

		response.Responses = append(response.Responses, models.Response{
			Href: href,
			Propstat: []models.Propstat{
				{
					Prop: models.Prop{
						GetETag:        etag,
						CalendarData:   vtodoContent,
						GetContentType: "text/calendar; charset=utf-8; component=vtodo",
					},
					Status: "HTTP/1.1 200 OK",
				},
			},
		})
	}

	return response, nil
}
