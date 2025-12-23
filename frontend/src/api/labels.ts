import { Label } from '@/models/label'
import { Request } from '../utils/api'
import { transport } from './transport'

type LabelsResponse = {
  labels: Label[]
}

type SingleLabelResponse = {
  label: Label
}

export const CreateLabel = async (label: Omit<Label, 'id'>) =>
  await transport({
    http: () => Request<SingleLabelResponse>(`/labels`, 'POST', label),
    ws: (ws) => ws.request('create_label', label),
  })

export const GetLabels = async () =>
  await transport({
    http: () => Request<LabelsResponse>(`/labels`),
    ws: (ws) => ws.request('get_user_labels'),
  })

export const UpdateLabel = async (label: Label) =>
  await transport({
    http: () => Request<SingleLabelResponse>(`/labels`, 'PUT', label),
    ws: (ws) => ws.request('update_label', label),
  })

export const DeleteLabel = async (id: number) =>
  await transport({
    http: () => Request<void>(`/labels/${id}`, 'DELETE'),
    ws: (ws) => ws.request('delete_label', id),
  })
