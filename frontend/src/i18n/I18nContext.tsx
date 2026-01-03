import React, { createContext, useContext, useState, useCallback, ReactNode, useEffect } from 'react';
import { Language, Translations, translations } from './translations';

interface I18nContextType {
  language: Language;
  setLanguage: (lang: Language) => void;
  t: Translations;
}

const I18nContext = createContext<I18nContextType | undefined>(undefined);

const STORAGE_KEY = 'vacation-planner-language';

export const useI18n = () => {
  const context = useContext(I18nContext);
  if (!context) {
    throw new Error('useI18n must be used within an I18nProvider');
  }
  return context;
};

// Helper hook for quick access to translations
export const useTranslations = () => {
  const { t } = useI18n();
  return t;
};

interface I18nProviderProps {
  children: ReactNode;
}

export const I18nProvider: React.FC<I18nProviderProps> = ({ children }) => {
  const [language, setLanguageState] = useState<Language>(() => {
    // Try to get from localStorage
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored === 'en' || stored === 'pt-PT') {
      return stored;
    }
    // Try to detect from browser
    const browserLang = navigator.language;
    if (browserLang.startsWith('pt')) {
      return 'pt-PT';
    }
    return 'en';
  });

  const setLanguage = useCallback((lang: Language) => {
    setLanguageState(lang);
    localStorage.setItem(STORAGE_KEY, lang);
  }, []);

  // Update document lang attribute
  useEffect(() => {
    document.documentElement.lang = language;
  }, [language]);

  const t = translations[language];

  return (
    <I18nContext.Provider value={{ language, setLanguage, t }}>
      {children}
    </I18nContext.Provider>
  );
};

// Utility function to interpolate strings with placeholders
// Usage: interpolate("Hello {0}, you have {1} messages", ["John", 5])
export const interpolate = (template: string, values: (string | number)[]): string => {
  return template.replace(/\{(\d+)\}/g, (match, index) => {
    const idx = parseInt(index, 10);
    return values[idx] !== undefined ? String(values[idx]) : match;
  });
};
