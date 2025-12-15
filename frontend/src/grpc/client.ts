import * as grpcWeb from 'grpc-web'
import { isTokenValid, refreshAccessToken } from '@/utils/api'
import { TOKEN_REFRESH_THRESHOLD_MS } from '@/constants/time'
import {
  TaskServiceDefinition,
  LabelServiceDefinition,
  UserServiceDefinition,
  TokenServiceDefinition,
  AuthServiceDefinition,
} from './task_service'
import type {
  TasksResponse,
  TaskResponse,
  TaskHistoryResponse,
} from './task'
import type {
  CreateTaskRequest,
  UpdateTaskRequest,
  GetTaskRequest,
  DeleteTaskRequest,
  CompleteTaskRequest,
  UncompleteTaskRequest,
  SkipTaskRequest,
  UpdateDueDateRequest,
  GetTaskHistoryRequest,
  GetCompletedTasksRequest,
} from './task'
import type { LabelsResponse, LabelResponse, CreateLabelRequest, UpdateLabelRequest, DeleteLabelRequest } from './label'
import type {
  UserProfileResponse,
  AppTokensResponse,
  AppTokenResponse,
  AuthResponse,
  SignUpRequest,
  LoginRequest,
  RefreshTokenRequest,
  ResetPasswordRequest,
  UpdatePasswordRequest,
  ChangePasswordRequest,
  CreateAppTokenRequest,
  DeleteAppTokenRequest,
  UpdateNotificationSettingsRequest,
} from './user'
import type { Empty } from './common'

const host = import.meta.env.VITE_GRPC_HOST || import.meta.env.VITE_APP_API_URL

type MetadataProvider = () => Promise<grpcWeb.Metadata>

function createGrpcWebClient<TDef extends { fullName: string; methods: Record<string, any> }>(
  serviceDef: TDef,
  hostname: string,
  metadataProvider?: MetadataProvider,
) {
  const client = new grpcWeb.GrpcWebClientBase({ format: 'text' })
  const methods: Record<string, any> = {}

  for (const [methodKey, methodDef] of Object.entries(serviceDef.methods)) {
    const methodDescriptor = new grpcWeb.MethodDescriptor(
      `/${serviceDef.fullName}/${methodDef.name}`,
      grpcWeb.MethodType.UNARY,
      Object as any,
      Object as any,
      (req: any) => methodDef.requestType.encode(req).finish(),
      (bytes: Uint8Array) => methodDef.responseType.decode(bytes),
    )

    methods[methodKey] = async (request: any) => {
      let metadata: grpcWeb.Metadata = {}
      if (metadataProvider) {
        metadata = await metadataProvider()
      }

      return new Promise((resolve, reject) => {
        client.rpcCall(
          hostname + `/${serviceDef.fullName}/${methodDef.name}`,
          request,
          metadata,
          methodDescriptor,
          (err: grpcWeb.RpcError, response: any) => {
            if (err) reject(err)
            else resolve(response)
          },
        )
      })
    }
  }

  return methods
}

const isTokenNearExpiration = () => {
  const now = new Date()
  const expiration = localStorage.getItem('ca_expiration') || ''
  const expire = new Date(expiration)
  return now.getTime() + TOKEN_REFRESH_THRESHOLD_MS > expire.getTime()
}

const authMetadataProvider: MetadataProvider = async () => {
  if (isTokenValid() && isTokenNearExpiration()) {
    await refreshAccessToken()
  }

  const token = localStorage.getItem('ca_token')
  if (!token) {
    throw new Error('User is not authenticated')
  }

  return {
    Authorization: `Bearer ${token}`,
  }
}

const taskServiceClient = createGrpcWebClient(TaskServiceDefinition, host, authMetadataProvider)
const labelServiceClient = createGrpcWebClient(LabelServiceDefinition, host, authMetadataProvider)
const userServiceClient = createGrpcWebClient(UserServiceDefinition, host, authMetadataProvider)
const tokenServiceClient = createGrpcWebClient(TokenServiceDefinition, host, authMetadataProvider)
const authServiceClient = createGrpcWebClient(AuthServiceDefinition, host)

// ============== Task Service ==============

export async function getTasks(): Promise<TasksResponse> {
  const request: Empty = {}
  return taskServiceClient.getTasks(request)
}

export async function getCompletedTasks(limit: number, page: number): Promise<TasksResponse> {
  const request: GetCompletedTasksRequest = { limit, page }
  return taskServiceClient.getCompletedTasks(request)
}

export async function getTask(id: number): Promise<TaskResponse> {
  const request: GetTaskRequest = { id }
  return taskServiceClient.getTask(request)
}

export async function createTask(task: CreateTaskRequest): Promise<TaskResponse> {
  return taskServiceClient.createTask(task)
}

export async function updateTask(task: UpdateTaskRequest): Promise<TaskResponse> {
  return taskServiceClient.updateTask(task)
}

export async function deleteTask(id: number): Promise<Empty> {
  const request: DeleteTaskRequest = { id }
  return taskServiceClient.deleteTask(request)
}

export async function completeTask(id: number, endRecurrence: boolean = false): Promise<TaskResponse> {
  const request: CompleteTaskRequest = { id, endRecurrence }
  return taskServiceClient.completeTask(request)
}

export async function uncompleteTask(id: number): Promise<TaskResponse> {
  const request: UncompleteTaskRequest = { id }
  return taskServiceClient.uncompleteTask(request)
}

export async function skipTask(id: number): Promise<TaskResponse> {
  const request: SkipTaskRequest = { id }
  return taskServiceClient.skipTask(request)
}

export async function updateDueDate(id: number, dueDate: number): Promise<TaskResponse> {
  const request: UpdateDueDateRequest = { id, dueDate }
  return taskServiceClient.updateDueDate(request)
}

export async function getTaskHistory(id: number): Promise<TaskHistoryResponse> {
  const request: GetTaskHistoryRequest = { id }
  return taskServiceClient.getTaskHistory(request)
}

// ============== Label Service ==============

export async function getLabels(): Promise<LabelsResponse> {
  const request: Empty = {}
  return labelServiceClient.getLabels(request)
}

export async function createLabel(label: CreateLabelRequest): Promise<LabelResponse> {
  return labelServiceClient.createLabel(label)
}

export async function updateLabel(label: UpdateLabelRequest): Promise<LabelResponse> {
  return labelServiceClient.updateLabel(label)
}

export async function deleteLabel(id: number): Promise<Empty> {
  const request: DeleteLabelRequest = { id }
  return labelServiceClient.deleteLabel(request)
}

// ============== User Service ==============

export async function getProfile(): Promise<UserProfileResponse> {
  const request: Empty = {}
  return userServiceClient.getProfile(request)
}

export async function updateNotificationSettings(
  settings: UpdateNotificationSettingsRequest,
): Promise<Empty> {
  return userServiceClient.updateNotificationSettings(settings)
}

export async function changePassword(newPassword: string): Promise<Empty> {
  const request: ChangePasswordRequest = { newPassword }
  return userServiceClient.changePassword(request)
}

// ============== Token Service ==============

export async function getAppTokens(): Promise<AppTokensResponse> {
  const request: Empty = {}
  return tokenServiceClient.getAppTokens(request)
}

export async function createAppToken(token: CreateAppTokenRequest): Promise<AppTokenResponse> {
  return tokenServiceClient.createAppToken(token)
}

export async function deleteAppToken(id: number): Promise<Empty> {
  const request: DeleteAppTokenRequest = { id }
  return tokenServiceClient.deleteAppToken(request)
}

// ============== Auth Service (unauthenticated) ==============

export async function signUp(request: SignUpRequest): Promise<Empty> {
  return authServiceClient.signUp(request)
}

export async function login(email: string, password: string): Promise<AuthResponse> {
  const request: LoginRequest = { email, password }
  return authServiceClient.login(request)
}

export async function refreshToken(refreshTokenValue: string): Promise<AuthResponse> {
  const request: RefreshTokenRequest = { refreshToken: refreshTokenValue }
  return authServiceClient.refreshToken(request)
}

export async function resetPassword(email: string): Promise<Empty> {
  const request: ResetPasswordRequest = { email }
  return authServiceClient.resetPassword(request)
}

export async function updatePassword(code: string, newPassword: string): Promise<Empty> {
  const request: UpdatePasswordRequest = { code, newPassword }
  return authServiceClient.updatePassword(request)
}
