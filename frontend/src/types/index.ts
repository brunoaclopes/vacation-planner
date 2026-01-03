export interface YearConfig {
  id: number;
  year: number;
  vacation_days: number;
  reserved_days: number;
  optimization_strategy: string;
  work_week: string[];
  optimizer_notes: string;
  created_at?: string;
  updated_at?: string;
}

export interface VacationDay {
  id: number;
  year: number;
  date: string;
  is_manual: boolean;
  note?: string;
  created_at?: string;
}

export interface OptimalVacation {
  id: number;
  year: number;
  date: string;
  block_id: number;
  consecutive_days: number;
  created_at?: string;
}

export interface Holiday {
  id?: number;
  year: number;
  date: string;
  name: string;
  type: string;
}

export interface CalendarDay {
  date: string;
  day_of_week: string;
  is_weekend: boolean;
  is_holiday: boolean;
  holiday_name?: string;
  is_vacation: boolean;
  is_manual: boolean;
  is_optimal: boolean;
  block_id?: number;
}

export interface VacationBlock {
  start_date: string;
  end_date: string;
  total_days: number;
  vacation_days_used: number;
  dates: string[];
  holidays: string[];
  weekends: string[];
}

export interface CalendarSummary {
  total_vacation_days: number;
  used_vacation_days: number;
  remaining_vacation_days: number;
  total_holidays: number;
  longest_vacation_block: number;
  total_days_off: number;
}

export interface CalendarResponse {
  year: number;
  config: YearConfig;
  days: CalendarDay[];
  holidays: Holiday[];
  vacation_blocks: VacationBlock[];
  manual_vacations: VacationDay[];
  optimal_vacations: OptimalVacation[];
  summary: CalendarSummary;
}

export interface ChatMessage {
  id: number;
  year: number;
  role: 'user' | 'assistant';
  content: string;
  created_at: string;
}

export interface Settings {
  openai_api_key: string;
  ai_provider: string;
  ai_model: string;
  backend_port: string;
  frontend_port: string;
  default_work_week: string;
  default_vacation_days: string;
  default_optimization_strategy: string;
  work_city: string;
  calendarific_api_key: string;
}

export interface OptimizationStrategy {
  id: string;
  name: string;
  description: string;
}

export const WORK_WEEK_PRESETS: Record<string, string[]> = {
  standard: ['monday', 'tuesday', 'wednesday', 'thursday', 'friday'],
  four_day: ['monday', 'tuesday', 'wednesday', 'thursday'],
  four_day_fri: ['tuesday', 'wednesday', 'thursday', 'friday'],
  six_day: ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday'],
  custom: [],
};

export const ALL_WEEKDAYS = [
  'monday',
  'tuesday',
  'wednesday',
  'thursday',
  'friday',
  'saturday',
  'sunday',
];
