import React, { createContext, useContext, useState, useCallback, ReactNode } from 'react';
import {
  CalendarResponse,
  YearConfig,
  ChatMessage,
} from '../types';
import * as api from '../services/api';
import { useI18n } from '../i18n';

interface CalendarContextType {
  year: number;
  setYear: (year: number) => void;
  calendar: CalendarResponse | null;
  loading: boolean;
  error: string | null;
  loadCalendar: (year: number) => Promise<void>;
  optimize: () => Promise<void>;
  clearOptimized: () => Promise<void>;
  addVacationDay: (date: string, note?: string) => Promise<void>;
  removeVacationDay: (date: string) => Promise<void>;
  updateConfig: (config: Partial<YearConfig>) => Promise<void>;
  chatMessages: ChatMessage[];
  sendMessage: (message: string) => Promise<void>;
  loadChatHistory: () => Promise<void>;
  clearChat: () => Promise<void>;
  chatLoading: boolean;
  // AI Suggestions
  suggestion: string | null;
  suggestionLoading: boolean;
  fetchSuggestions: () => Promise<void>;
}

const CalendarContext = createContext<CalendarContextType | undefined>(undefined);

export const useCalendar = () => {
  const context = useContext(CalendarContext);
  if (!context) {
    throw new Error('useCalendar must be used within a CalendarProvider');
  }
  return context;
};

interface CalendarProviderProps {
  children: ReactNode;
}

export const CalendarProvider: React.FC<CalendarProviderProps> = ({ children }) => {
  const [year, setYear] = useState<number>(new Date().getFullYear());
  const [calendar, setCalendar] = useState<CalendarResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [chatLoading, setChatLoading] = useState(false);
  
  // AI Suggestions state
  const [suggestion, setSuggestion] = useState<string | null>(null);
  const [suggestionLoading, setSuggestionLoading] = useState(false);
  const [suggestionCacheKey, setSuggestionCacheKey] = useState<string | null>(null);
  
  // Get current language from i18n context
  const { language } = useI18n();

  // Generate a cache key based on calendar state that affects suggestions
  const getCalendarFingerprint = useCallback((cal: CalendarResponse | null, yr: number, lang: string): string => {
    if (!cal) return '';
    return `${yr}-${lang}-${cal.summary?.used_vacation_days ?? 0}-${cal.summary?.total_days_off ?? 0}-${cal.optimal_vacations?.length ?? 0}-${cal.config?.vacation_days ?? 0}`;
  }, []);

  const loadCalendar = useCallback(async (targetYear: number) => {
    setLoading(true);
    setError(null);
    try {
      const data = await api.getCalendar(targetYear);
      setCalendar(data);
      setYear(targetYear);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load calendar');
    } finally {
      setLoading(false);
    }
  }, []);

  const optimize = useCallback(async () => {
    setLoading(true);
    try {
      await api.optimizeVacations(year);
      await loadCalendar(year);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to optimize');
    } finally {
      setLoading(false);
    }
  }, [year, loadCalendar]);

  const clearOptimized = useCallback(async () => {
    setLoading(true);
    try {
      await api.clearOptimizedVacations(year);
      await loadCalendar(year);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to clear optimized days');
    } finally {
      setLoading(false);
    }
  }, [year, loadCalendar]);

  const addVacationDay = useCallback(async (date: string, note?: string) => {
    try {
      await api.addVacation(year, date, note);
      await loadCalendar(year);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to add vacation');
    }
  }, [year, loadCalendar]);

  const removeVacationDay = useCallback(async (date: string) => {
    try {
      await api.removeVacation(year, date);
      await loadCalendar(year);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to remove vacation');
    }
  }, [year, loadCalendar]);

  const updateConfig = useCallback(async (config: Partial<YearConfig>) => {
    try {
      await api.updateYearConfig(year, config);
      await loadCalendar(year);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update config');
    }
  }, [year, loadCalendar]);

  const loadChatHistory = useCallback(async () => {
    try {
      const history = await api.getChatHistory(year);
      setChatMessages(history || []);
    } catch (err) {
      console.error('Failed to load chat history:', err);
    }
  }, [year]);

  const sendMessage = useCallback(async (message: string) => {
    setChatLoading(true);
    try {
      // Add user message immediately
      const userMessage: ChatMessage = {
        id: Date.now(),
        year,
        role: 'user',
        content: message,
        created_at: new Date().toISOString(),
      };
      setChatMessages(prev => [...prev, userMessage]);

      const response = await api.sendChatMessage(year, message);
      
      // Add assistant message
      const assistantMessage: ChatMessage = {
        id: Date.now() + 1,
        year,
        role: 'assistant',
        content: response.message,
        created_at: new Date().toISOString(),
      };
      setChatMessages(prev => [...prev, assistantMessage]);

      // If there was an action, refresh the calendar
      if (response.hasAction) {
        if (response.action?.triggerOptimize) {
          await optimize();
        } else {
          await loadCalendar(year);
        }
      }
    } catch (err) {
      const errorMessage: ChatMessage = {
        id: Date.now() + 1,
        year,
        role: 'assistant',
        content: `Error: ${err instanceof Error ? err.message : 'Failed to send message'}`,
        created_at: new Date().toISOString(),
      };
      setChatMessages(prev => [...prev, errorMessage]);
    } finally {
      setChatLoading(false);
    }
  }, [year, loadCalendar, optimize]);

  const clearChat = useCallback(async () => {
    try {
      await api.clearChatHistory(year);
      setChatMessages([]);
    } catch (err) {
      console.error('Failed to clear chat:', err);
    }
  }, [year]);

  // Fetch AI suggestions with caching
  const fetchSuggestions = useCallback(async () => {
    const currentFingerprint = getCalendarFingerprint(calendar, year, language);
    
    // Skip if we already have cached suggestions for this calendar state
    if (suggestionCacheKey === currentFingerprint && suggestion !== null) {
      return;
    }
    
    setSuggestionLoading(true);
    try {
      const result = await api.getVacationSuggestions(year, language);
      setSuggestion(result.suggestion);
      setSuggestionCacheKey(currentFingerprint);
    } catch {
      setSuggestion(null);
    } finally {
      setSuggestionLoading(false);
    }
  }, [year, language, calendar, suggestionCacheKey, suggestion, getCalendarFingerprint]);

  return (
    <CalendarContext.Provider
      value={{
        year,
        setYear,
        calendar,
        loading,
        error,
        loadCalendar,
        optimize,
        clearOptimized,
        addVacationDay,
        removeVacationDay,
        updateConfig,
        chatMessages,
        sendMessage,
        loadChatHistory,
        clearChat,
        chatLoading,
        suggestion,
        suggestionLoading,
        fetchSuggestions,
      }}
    >
      {children}
    </CalendarContext.Provider>
  );
};
