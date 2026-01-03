export type Language = 'en' | 'pt-PT';

export interface Translations {
  // Common
  common: {
    cancel: string;
    save: string;
    clear: string;
    loading: string;
    error: string;
    success: string;
    settings: string;
    calendar: string;
    year: string;
    days: string;
    day: string;
    remaining: string;
    used: string;
    of: string;
  };

  // Menu
  menu: {
    calendar: string;
    settings: string;
  };

  // App
  app: {
    title: string;
    subtitle: string;
    copyright: string;
  };

  // Theme
  theme: {
    light: string;
    dark: string;
    system: string;
  };

  // Calendar
  calendar: {
    title: string;
    weekdays: string[];
    weekdaysShort: string[];
    months: string[];
    monthsShort: string[];
    vacation: string;
    holiday: string;
    weekend: string;
    manualVacation: string;
    optimizedVacation: string;
    optimizeVacations: string;
    addVacation: string;
    removeVacation: string;
    clearOptimized: string;
    aiAssistant: string;
    gettingAiSuggestions: string;
    dayDetails: string;
    clickToAdd: string;
    clickToRemove: string;
    isAHoliday: string;
    tooltipManualVacation: string;
    tooltipOptimizedVacation: string;
    holidaysListTitle: string;
    dayIsVacation: string;
    dayIsVacationManual: string;
    dayIsVacationOptimized: string;
    dayIsWorkday: string;
  };

  // Summary
  summary: {
    title: string;
    vacationDaysUsed: string;
    holidaysThisYear: string;
    longestBlock: string;
    totalDaysOff: string;
    vacationDaysUsage: string;
  };

  // Chat
  chat: {
    title: string;
    emptyState: string;
    emptyStateHint: string;
    placeholder: string;
  };

  // Year Config
  config: {
    title: string;
    totalVacationDays: string;
    totalVacationDaysHelp: string;
    reservedDays: string;
    reservedDaysHelp: string;
    availableForPlanning: string;
    optimizationStrategy: string;
    smartOptimizerNotes: string;
    smartOptimizerNotesPlaceholder: string;
    smartOptimizerNotesHelp: string;
    workWeekConfig: string;
    preset: string;
    presetStandard: string;
    preset4DayMonThu: string;
    preset4DayTueFri: string;
    preset6Day: string;
    presetCustom: string;
    saveConfig: string;
    configSaved: string;
  };

  // Settings
  settings: {
    title: string;
    aiIntegration: string;
    aiIntegrationDesc: string;
    aiProvider: string;
    providerGitHub: string;
    providerOpenAI: string;
    githubToken: string;
    openaiKey: string;
    githubTokenPlaceholder: string;
    openaiKeyPlaceholder: string;
    githubTokenHelp: string;
    openaiKeyHelp: string;
    aiModel: string;
    refreshModels: string;
    defaultYearConfig: string;
    defaultYearConfigDesc: string;
    defaultVacationDays: string;
    defaultStrategy: string;
    strategyBridgeHolidays: string;
    strategyLongestBlocks: string;
    strategyBalanced: string;
    defaultWorkWeek: string;
    locationSettings: string;
    locationSettingsDesc: string;
    workCity: string;
    workCityPlaceholder: string;
    workCityHelp: string;
    calendarificKey: string;
    calendarificKeyPlaceholder: string;
    calendarificKeyHelp: string;
    howToUse: string;
    howToUseStep1: string;
    howToUseStep1Desc: string;
    howToUseStep2: string;
    howToUseStep2Desc: string;
    howToUseStep3: string;
    howToUseStep3Desc: string;
    howToUseStep4: string;
    howToUseStep4Desc: string;
    howToUseStep5: string;
    howToUseStep5Desc: string;
    saveAllSettings: string;
    settingsSaved: string;
    language: string;
    languageDesc: string;
  };

  // Holidays
  holidays: {
    errorTitle: string;
    errorDesc: string;
    retryNow: string;
    hideDetails: string;
    showDetails: string;
    dismiss: string;
    failedToLoad: string;
    retrying: string;
    maxRetriesReached: string;
  };

  // Errors
  errors: {
    failedToLoad: string;
    failedToOptimize: string;
    failedToClear: string;
    failedToAdd: string;
    failedToRemove: string;
    failedToUpdate: string;
    failedToSend: string;
    failedToSave: string;
  };
}
