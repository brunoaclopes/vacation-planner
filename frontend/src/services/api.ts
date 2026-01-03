import axios from 'axios';
import {
  CalendarResponse,
  YearConfig,
  VacationDay,
  Holiday,
  ChatMessage,
  Settings,
  OptimizationStrategy,
  VacationBlock,
} from '../types';

const api = axios.create({
  baseURL: '/api',
});

// Calendar
export const getCalendar = async (year: number): Promise<CalendarResponse> => {
  const response = await api.get<CalendarResponse>(`/calendar/${year}`);
  return response.data;
};

export const optimizeVacations = async (
  year: number
): Promise<{ blocks: VacationBlock[]; message: string }> => {
  const response = await api.post(`/calendar/${year}/optimize`);
  return response.data;
};

export const clearOptimizedVacations = async (year: number): Promise<void> => {
  await api.delete(`/calendar/${year}/optimized`);
};

export const getVacationSuggestions = async (year: number, language: string = 'en'): Promise<{ suggestion: string }> => {
  const response = await api.get<{ suggestion: string }>(`/calendar/${year}/suggestions`, {
    params: { language }
  });
  return response.data;
};

// Vacations
export const getVacations = async (year: number): Promise<VacationDay[]> => {
  const response = await api.get<VacationDay[]>(`/vacations/${year}`);
  return response.data;
};

export const addVacation = async (
  year: number,
  date: string,
  note?: string
): Promise<void> => {
  await api.post(`/vacations/${year}`, { date, note });
};

export const removeVacation = async (
  year: number,
  date: string
): Promise<void> => {
  await api.delete(`/vacations/${year}/${date}`);
};

export const bulkUpdateVacations = async (
  year: number,
  add: string[],
  remove: string[]
): Promise<void> => {
  await api.put(`/vacations/${year}/bulk`, { add, remove });
};

// Holidays
export const getHolidays = async (year: number): Promise<Holiday[]> => {
  const response = await api.get<Holiday[]>(`/holidays/${year}`);
  return response.data;
};

export const refreshHolidays = async (year: number): Promise<{ message: string; holidays: Holiday[]; has_errors?: boolean; status?: HolidayStatus }> => {
  const response = await api.post<{ message: string; holidays: Holiday[]; has_errors?: boolean; status?: HolidayStatus }>(`/holidays/${year}/refresh`);
  return response.data;
};

// Holiday status
export interface HolidayStatus {
  year: number;
  national_loaded: boolean;
  municipal_loaded: boolean;
  national_error?: string;
  municipal_error?: string;
  last_updated: string;
  retry_count: number;
  max_retries: number;
  is_retrying: boolean;
  next_retry?: string;
  has_errors: boolean;
}

export const getHolidayStatus = async (year: number): Promise<HolidayStatus> => {
  const response = await api.get<HolidayStatus>(`/holidays/${year}/status`);
  return response.data;
};

export const getAllHolidayStatuses = async (): Promise<HolidayStatus[]> => {
  const response = await api.get<HolidayStatus[]>(`/holidays/status`);
  return response.data;
};

// Year Config
export const getYearConfig = async (year: number): Promise<YearConfig> => {
  const response = await api.get<YearConfig>(`/config/${year}`);
  return response.data;
};

export const updateYearConfig = async (
  year: number,
  config: Partial<YearConfig>
): Promise<YearConfig> => {
  const response = await api.put<YearConfig>(`/config/${year}`, config);
  return response.data;
};

export const copyYearConfig = async (
  year: number,
  sourceYear: number
): Promise<void> => {
  await api.post(`/config/${year}/copy-from/${sourceYear}`);
};

// Settings
export const getSettings = async (): Promise<Settings> => {
  const response = await api.get<Settings>('/settings');
  return response.data;
};

export const updateSettings = async (
  settings: Partial<Settings>
): Promise<void> => {
  await api.put('/settings', settings);
};

export const getSetting = async (key: string): Promise<string> => {
  const response = await api.get<Record<string, string>>(`/settings/${key}`);
  return response.data[key];
};

export const updateSetting = async (
  key: string,
  value: string
): Promise<void> => {
  await api.put(`/settings/${key}`, { value });
};

// Chat
export const sendChatMessage = async (
  year: number,
  message: string
): Promise<{
  message: string;
  action: Record<string, unknown> | null;
  hasAction: boolean;
}> => {
  const response = await api.post(`/chat/${year}`, { message });
  return response.data;
};

export const getChatHistory = async (year: number): Promise<ChatMessage[]> => {
  const response = await api.get<ChatMessage[]>(`/chat/${year}/history`);
  return response.data;
};

export const clearChatHistory = async (year: number): Promise<void> => {
  await api.delete(`/chat/${year}/history`);
};

// AI Models
export interface AIModel {
  id: string;
  name: string;
  publisher: string;
}

export const getAvailableModels = async (): Promise<AIModel[]> => {
  const response = await api.get<AIModel[]>('/models');
  return response.data;
};

// Presets
export const getWorkWeekPresets = async (): Promise<
  Record<string, string[]>
> => {
  const response = await api.get<Record<string, string[]>>(
    '/presets/work-week'
  );
  return response.data;
};

export const getOptimizationStrategies = async (): Promise<
  OptimizationStrategy[]
> => {
  const response = await api.get<OptimizationStrategy[]>('/presets/strategies');
  return response.data;
};

// Cities for municipal holidays
export const getAvailableCities = async (): Promise<string[]> => {
  const response = await api.get<string[]>('/cities');
  return response.data;
};

// Version
export const getVersion = async (): Promise<string> => {
  try {
    const response = await api.get<{ version: string }>('/version');
    return response.data.version;
  } catch {
    return import.meta.env.VITE_APP_VERSION || 'dev';
  }
};
