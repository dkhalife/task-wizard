import { Label } from '@/models/label'
import { Request } from '../utils/api'

type LabelsResponse = {
  labels: Label[]
}

type SingleLabelResponse = {
  label: Label
}

export const CreateLabel = async (label: Omit<Label, 'id'>) =>
  await Request<SingleLabelResponse>(`/labels`, 'POST', label)
export const GetLabels = async () => await Request<LabelsResponse>(`/labels`)
export const UpdateLabel = async (label: Label) =>
  await Request<SingleLabelResponse>(`/labels`, 'PUT', label)
export const DeleteLabel = async (id: number) =>
  await Request<void>(`/labels/${id}`, 'DELETE')
