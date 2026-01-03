import React, { createContext, useContext, useState, useEffect, useMemo } from 'react';
import { ThemeProvider as MuiThemeProvider, createTheme } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import useMediaQuery from '@mui/material/useMediaQuery';

type PaletteMode = 'light' | 'dark';

type ThemeMode = 'light' | 'dark' | 'system';

interface ThemeContextType {
  mode: ThemeMode;
  setMode: (mode: ThemeMode) => void;
  resolvedMode: PaletteMode;
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

export const useThemeMode = () => {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error('useThemeMode must be used within a ThemeProvider');
  }
  return context;
};

const getDesignTokens = (mode: PaletteMode) => ({
  palette: {
    mode,
    ...(mode === 'light'
      ? {
          // Light mode colors
          primary: {
            main: '#1a5f2a',
            light: '#2e7d32',
            dark: '#0d4a1a',
          },
          secondary: {
            main: '#b71c1c',
            light: '#e53935',
            dark: '#7f0000',
          },
          background: {
            default: '#f8fafc',
            paper: '#ffffff',
          },
          text: {
            primary: '#1e293b',
            secondary: '#64748b',
          },
          grey: {
            50: '#f8fafc',
            100: '#f1f5f9',
            200: '#e2e8f0',
            300: '#cbd5e1',
            400: '#94a3b8',
            500: '#64748b',
            600: '#475569',
            700: '#334155',
            800: '#1e293b',
            900: '#0f172a',
          },
          success: {
            main: '#059669',
            light: '#10b981',
          },
          info: {
            main: '#0284c7',
            light: '#38bdf8',
          },
          warning: {
            main: '#d97706',
            light: '#f59e0b',
          },
          error: {
            main: '#b71c1c',
            light: '#ef4444',
          },
          // Custom semantic colors
          vacation: {
            main: '#1a5f2a',
            light: '#22c55e',
          },
          holiday: {
            main: '#ef4444',
            light: '#f87171',
          },
          optimized: {
            main: '#0d9488',
            light: '#14b8a6',
          },
          divider: '#e2e8f0',
        }
      : {
          // Dark mode colors
          primary: {
            main: '#16a34a',
            light: '#22c55e',
            dark: '#15803d',
            contrastText: '#ffffff',
          },
          secondary: {
            main: '#f87171',
            light: '#fca5a5',
            dark: '#ef4444',
          },
          background: {
            default: '#0f172a',
            paper: '#1e293b',
          },
          text: {
            primary: '#f1f5f9',
            secondary: '#94a3b8',
          },
          grey: {
            50: '#0f172a',
            100: '#1e293b',
            200: '#334155',
            300: '#475569',
            400: '#64748b',
            500: '#94a3b8',
            600: '#cbd5e1',
            700: '#e2e8f0',
            800: '#f1f5f9',
            900: '#f8fafc',
          },
          success: {
            main: '#10b981',
            light: '#34d399',
          },
          info: {
            main: '#38bdf8',
            light: '#7dd3fc',
          },
          warning: {
            main: '#f59e0b',
            light: '#fbbf24',
          },
          error: {
            main: '#ef4444',
            light: '#f87171',
          },
          // Custom semantic colors
          vacation: {
            main: '#22c55e',
            light: '#4ade80',
          },
          holiday: {
            main: '#f87171',
            light: '#fca5a5',
          },
          optimized: {
            main: '#14b8a6',
            light: '#2dd4bf',
          },
          divider: '#334155',
        }),
  },
  typography: {
    fontFamily: '"Inter", "Roboto", "Helvetica", "Arial", sans-serif',
    h1: {
      fontWeight: 700,
      letterSpacing: '-0.02em',
    },
    h2: {
      fontWeight: 700,
      letterSpacing: '-0.01em',
    },
    h3: {
      fontWeight: 600,
      letterSpacing: '-0.01em',
    },
    h4: {
      fontWeight: 600,
    },
    h5: {
      fontWeight: 600,
    },
    h6: {
      fontWeight: 600,
    },
    subtitle1: {
      fontWeight: 500,
    },
    subtitle2: {
      fontWeight: 500,
    },
    body1: {
      lineHeight: 1.6,
    },
    body2: {
      lineHeight: 1.5,
    },
    button: {
      fontWeight: 600,
      textTransform: 'none' as const,
    },
  },
  shape: {
    borderRadius: 12,
  },
  shadows: [
    'none',
    mode === 'light' ? '0 1px 2px 0 rgb(0 0 0 / 0.05)' : '0 1px 2px 0 rgb(0 0 0 / 0.3)',
    mode === 'light' ? '0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1)' : '0 1px 3px 0 rgb(0 0 0 / 0.4), 0 1px 2px -1px rgb(0 0 0 / 0.3)',
    mode === 'light' ? '0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)' : '0 4px 6px -1px rgb(0 0 0 / 0.4), 0 2px 4px -2px rgb(0 0 0 / 0.3)',
    mode === 'light' ? '0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1)' : '0 10px 15px -3px rgb(0 0 0 / 0.4), 0 4px 6px -4px rgb(0 0 0 / 0.3)',
    mode === 'light' ? '0 20px 25px -5px rgb(0 0 0 / 0.1), 0 8px 10px -6px rgb(0 0 0 / 0.1)' : '0 20px 25px -5px rgb(0 0 0 / 0.4), 0 8px 10px -6px rgb(0 0 0 / 0.3)',
    '0 25px 50px -12px rgb(0 0 0 / 0.25)',
    '0 25px 50px -12px rgb(0 0 0 / 0.25)',
    '0 25px 50px -12px rgb(0 0 0 / 0.25)',
    '0 25px 50px -12px rgb(0 0 0 / 0.25)',
    '0 25px 50px -12px rgb(0 0 0 / 0.25)',
    '0 25px 50px -12px rgb(0 0 0 / 0.25)',
    '0 25px 50px -12px rgb(0 0 0 / 0.25)',
    '0 25px 50px -12px rgb(0 0 0 / 0.25)',
    '0 25px 50px -12px rgb(0 0 0 / 0.25)',
    '0 25px 50px -12px rgb(0 0 0 / 0.25)',
    '0 25px 50px -12px rgb(0 0 0 / 0.25)',
    '0 25px 50px -12px rgb(0 0 0 / 0.25)',
    '0 25px 50px -12px rgb(0 0 0 / 0.25)',
    '0 25px 50px -12px rgb(0 0 0 / 0.25)',
    '0 25px 50px -12px rgb(0 0 0 / 0.25)',
    '0 25px 50px -12px rgb(0 0 0 / 0.25)',
    '0 25px 50px -12px rgb(0 0 0 / 0.25)',
    '0 25px 50px -12px rgb(0 0 0 / 0.25)',
    '0 25px 50px -12px rgb(0 0 0 / 0.25)',
  ] as const,
  components: {
    MuiButton: {
      styleOverrides: {
        root: {
          borderRadius: 8,
          padding: '8px 16px',
          boxShadow: 'none',
          '&:hover': {
            boxShadow: mode === 'light' 
              ? '0 4px 6px -1px rgb(0 0 0 / 0.1)' 
              : '0 4px 6px -1px rgb(0 0 0 / 0.4)',
          },
        },
        contained: {
          '&:hover': {
            boxShadow: mode === 'light' 
              ? '0 4px 6px -1px rgb(0 0 0 / 0.1)' 
              : '0 4px 6px -1px rgb(0 0 0 / 0.4)',
          },
        },
      },
    },
    MuiPaper: {
      styleOverrides: {
        root: {
          backgroundImage: 'none',
        },
        elevation1: {
          boxShadow: mode === 'light' 
            ? '0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1)' 
            : '0 1px 3px 0 rgb(0 0 0 / 0.4), 0 1px 2px -1px rgb(0 0 0 / 0.3)',
        },
        elevation2: {
          boxShadow: mode === 'light' 
            ? '0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)' 
            : '0 4px 6px -1px rgb(0 0 0 / 0.4), 0 2px 4px -2px rgb(0 0 0 / 0.3)',
        },
      },
    },
    MuiCard: {
      styleOverrides: {
        root: {
          borderRadius: 16,
          boxShadow: mode === 'light' 
            ? '0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)' 
            : '0 4px 6px -1px rgb(0 0 0 / 0.4), 0 2px 4px -2px rgb(0 0 0 / 0.3)',
        },
      },
    },
    MuiAppBar: {
      styleOverrides: {
        root: {
          boxShadow: mode === 'light' 
            ? '0 1px 3px 0 rgb(0 0 0 / 0.1)' 
            : '0 1px 3px 0 rgb(0 0 0 / 0.4)',
        },
      },
    },
    MuiChip: {
      styleOverrides: {
        root: {
          fontWeight: 500,
        },
      },
    },
    MuiLinearProgress: {
      styleOverrides: {
        root: {
          borderRadius: 4,
          backgroundColor: mode === 'light' ? '#e2e8f0' : '#334155',
        },
      },
    },
    MuiDrawer: {
      styleOverrides: {
        paper: {
          backgroundColor: mode === 'light' ? '#ffffff' : '#1e293b',
        },
      },
    },
    MuiDialog: {
      styleOverrides: {
        paper: {
          backgroundColor: mode === 'light' ? '#ffffff' : '#1e293b',
        },
      },
    },
    MuiMenu: {
      styleOverrides: {
        paper: {
          backgroundColor: mode === 'light' ? '#ffffff' : '#1e293b',
        },
      },
    },
    MuiTooltip: {
      styleOverrides: {
        tooltip: {
          backgroundColor: mode === 'light' ? '#1e293b' : '#f1f5f9',
          color: mode === 'light' ? '#f1f5f9' : '#1e293b',
        },
      },
    },
  },
});

interface ThemeProviderProps {
  children: React.ReactNode;
}

export const ThemeProvider: React.FC<ThemeProviderProps> = ({ children }) => {
  const prefersDarkMode = useMediaQuery('(prefers-color-scheme: dark)');
  const [mode, setMode] = useState<ThemeMode>(() => {
    const savedMode = localStorage.getItem('themeMode') as ThemeMode | null;
    return savedMode || 'system';
  });

  useEffect(() => {
    localStorage.setItem('themeMode', mode);
  }, [mode]);

  const resolvedMode: PaletteMode = useMemo(() => {
    if (mode === 'system') {
      return prefersDarkMode ? 'dark' : 'light';
    }
    return mode;
  }, [mode, prefersDarkMode]);

  const theme = useMemo(() => {
    const tokens = getDesignTokens(resolvedMode);
    return createTheme(tokens as any);
  }, [resolvedMode]);

  const contextValue = useMemo(
    () => ({ mode, setMode, resolvedMode }),
    [mode, resolvedMode]
  );

  return (
    <ThemeContext.Provider value={contextValue}>
      <MuiThemeProvider theme={theme}>
        <CssBaseline />
        {children}
      </MuiThemeProvider>
    </ThemeContext.Provider>
  );
};
