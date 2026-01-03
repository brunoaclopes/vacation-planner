import { useEffect, useState } from 'react';
import {
  Box,
  Typography,
  Button,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Grid,
  CircularProgress,
  Alert,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  IconButton,
  Drawer,
  useMediaQuery,
  useTheme,
} from '@mui/material';
import {
  ChevronLeft as ChevronLeftIcon,
  ChevronRight as ChevronRightIcon,
  AutoFixHigh as OptimizeIcon,
  Chat as ChatIcon,
  Close as CloseIcon,
  Lightbulb as SuggestionIcon,
} from '@mui/icons-material';
import { useCalendar } from '../context/CalendarContext';
import { useTranslations, useI18n, interpolate } from '../i18n';
import YearCalendar from '../components/YearCalendar';
import CalendarSummary from '../components/CalendarSummary';
import YearConfigPanel from '../components/YearConfigPanel';
import ChatPanel from '../components/ChatPanel';
import HolidayNotification from '../components/HolidayNotification';
import { CalendarDay } from '../types';

const CalendarPage: React.FC = () => {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const t = useTranslations();
  const { language } = useI18n();
  const {
    year,
    calendar,
    loading,
    error,
    loadCalendar,
    optimize,
    addVacationDay,
    removeVacationDay,
    suggestion,
    suggestionLoading,
    fetchSuggestions,
  } = useCalendar();

  const [selectedDay, setSelectedDay] = useState<CalendarDay | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [chatOpen, setChatOpen] = useState(false);

  // Generate year options (2025 to current year + 5)
  const currentYear = new Date().getFullYear();
  const yearOptions = Array.from({ length: currentYear - 2025 + 6 }, (_, i) => 2025 + i);

  useEffect(() => {
    loadCalendar(year);
  }, [year, loadCalendar]);

  // Auto-fetch suggestions when calendar data changes (with caching in context)
  useEffect(() => {
    if (calendar?.summary?.used_vacation_days !== undefined) {
      const timer = setTimeout(() => {
        fetchSuggestions();
      }, 500); // Small debounce to avoid too many requests
      return () => clearTimeout(timer);
    }
  }, [calendar?.summary?.used_vacation_days, calendar?.summary?.total_days_off, fetchSuggestions]);

  const handleYearChange = (newYear: number) => {
    loadCalendar(newYear);
  };

  const handleDayClick = (day: CalendarDay) => {
    // Don't allow clicking on weekends or holidays (they're already off)
    if (day.is_weekend || day.is_holiday) return;
    setSelectedDay(day);
    setDialogOpen(true);
  };

  const handleToggleVacation = async () => {
    if (!selectedDay) return;
    
    if (selectedDay.is_manual) {
      await removeVacationDay(selectedDay.date);
    } else {
      await addVacationDay(selectedDay.date);
    }
    setDialogOpen(false);
    setSelectedDay(null);
  };

  const handleOptimize = async () => {
    await optimize();
  };

  if (loading && !calendar) {
    return (
      <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '50vh' }}>
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box>
      {/* Holiday Status Notification */}
      <HolidayNotification year={year} onRefresh={() => loadCalendar(year)} />
      
      {/* Header Controls */}
      <Box 
        sx={{ 
          display: 'flex', 
          justifyContent: 'space-between', 
          alignItems: 'center', 
          mb: 4, 
          flexWrap: 'wrap', 
          gap: 2,
          pb: 3,
          borderBottom: '1px solid',
          borderColor: 'divider',
        }}
      >
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <IconButton 
            onClick={() => handleYearChange(year - 1)} 
            disabled={year <= 2025}
            sx={{ 
              bgcolor: isDark ? 'grey.200' : 'grey.100',
              color: 'text.primary',
              '&:hover': { bgcolor: isDark ? 'grey.300' : 'grey.200' },
              '&.Mui-disabled': { opacity: 0.5 },
            }}
          >
            <ChevronLeftIcon />
          </IconButton>
          
          <FormControl size="small" sx={{ minWidth: 120 }}>
            <InputLabel>{t.common.year}</InputLabel>
            <Select
              value={year}
              label={t.common.year}
              onChange={(e) => handleYearChange(e.target.value as number)}
              sx={{ fontWeight: 600 }}
            >
              {yearOptions.map((y) => (
                <MenuItem key={y} value={y}>{y}</MenuItem>
              ))}
            </Select>
          </FormControl>
          
          <IconButton 
            onClick={() => handleYearChange(year + 1)}
            sx={{ 
              bgcolor: isDark ? 'grey.200' : 'grey.100',
              color: 'text.primary',
              '&:hover': { bgcolor: isDark ? 'grey.300' : 'grey.200' },
            }}
          >
            <ChevronRightIcon />
          </IconButton>
        </Box>

        <Box sx={{ display: 'flex', gap: 1.5 }}>
          <Button
            variant="contained"
            startIcon={<OptimizeIcon />}
            onClick={handleOptimize}
            disabled={loading}
            sx={{ px: 3 }}
          >
            {t.calendar.optimizeVacations}
          </Button>
          <Button
            variant="outlined"
            startIcon={<ChatIcon />}
            onClick={() => setChatOpen(true)}
            sx={{ px: 3 }}
          >
            {t.calendar.aiAssistant}
          </Button>
        </Box>
      </Box>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      {calendar && (
        <Grid container spacing={3}>
          <Grid item xs={12} lg={9}>
            {/* Summary */}
            <CalendarSummary summary={calendar.summary} />
            
            {/* Calendar */}
            <YearCalendar
              year={year}
              days={calendar.days || []}
              onDayClick={handleDayClick}
            />

            {/* AI Suggestions */}
            <Box sx={{ mt: 2, px: 1 }}>
              {suggestionLoading ? (
                <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                  <CircularProgress size={14} sx={{ color: 'text.disabled' }} />
                  <Typography variant="body2" sx={{ color: 'text.disabled', fontStyle: 'italic' }}>
                    {t.calendar.gettingAiSuggestions}
                  </Typography>
                </Box>
              ) : suggestion ? (
                <Box sx={{ display: 'flex', gap: 1, alignItems: 'flex-start' }}>
                  <SuggestionIcon sx={{ fontSize: 16, color: 'text.disabled', mt: 0.3, flexShrink: 0 }} />
                  <Box
                    sx={{
                      color: 'text.disabled',
                      fontSize: '0.875rem',
                      lineHeight: 1.6,
                      '& strong': { fontWeight: 600 },
                      '& p': { my: 0.5 },
                      '& ul, & ol': { pl: 2, my: 0.5 },
                      '& li': { mb: 0.5 },
                    }}
                    dangerouslySetInnerHTML={{
                      __html: suggestion
                        // Escape HTML first
                        .replace(/</g, '&lt;')
                        .replace(/>/g, '&gt;')
                        // Headers
                        .replace(/^### (.+)$/gm, '<strong>$1</strong>')
                        .replace(/^## (.+)$/gm, '<strong>$1</strong>')
                        .replace(/^# (.+)$/gm, '<strong>$1</strong>')
                        // Bold
                        .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
                        // Italic
                        .replace(/\*(.+?)\*/g, '<em>$1</em>')
                        // Line breaks
                        .replace(/\n/g, '<br />')
                    }}
                  />
                </Box>
              ) : null}
            </Box>
          </Grid>

          <Grid item xs={12} lg={3}>
            {/* Configuration Panel */}
            <YearConfigPanel />

            {/* Holidays List */}
            <Box sx={{ mt: 2 }}>
              <Typography variant="h6" gutterBottom>
                {t.calendar.holidaysListTitle} {year}
              </Typography>
              {calendar.holidays?.map((holiday) => (
                <Box
                  key={holiday.date}
                  sx={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    py: 0.5,
                    borderBottom: '1px solid',
                    borderColor: 'divider',
                  }}
                >
                  <Typography variant="body2">{holiday.name}</Typography>
                  <Typography variant="body2" color="text.secondary">
                    {new Date(holiday.date).toLocaleDateString(language, {
                      month: 'short',
                      day: 'numeric',
                    })}
                  </Typography>
                </Box>
              ))}
            </Box>
          </Grid>
        </Grid>
      )}

      {/* Day Detail Dialog */}
      <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)}>
        <DialogTitle>
          {selectedDay && new Date(selectedDay.date).toLocaleDateString(language, {
            weekday: 'long',
            year: 'numeric',
            month: 'long',
            day: 'numeric',
          })}
        </DialogTitle>
        <DialogContent>
          {selectedDay && (
            <Box>
              {selectedDay.is_vacation ? (
                <Typography>
                  {t.calendar.dayIsVacation}
                  {selectedDay.is_manual && ` ${t.calendar.dayIsVacationManual}`}
                  {selectedDay.is_optimal && ` ${interpolate(t.calendar.dayIsVacationOptimized, [selectedDay.block_id || 0])}`}
                </Typography>
              ) : (
                <Typography>
                  {t.calendar.dayIsWorkday}
                </Typography>
              )}
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setDialogOpen(false)}>{t.common.cancel}</Button>
          <Button
            onClick={handleToggleVacation}
            variant="contained"
            color={selectedDay?.is_manual ? 'error' : 'primary'}
          >
            {selectedDay?.is_manual ? t.calendar.removeVacation : t.calendar.addVacation}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Chat Drawer */}
      <Drawer
        anchor={isMobile ? 'bottom' : 'right'}
        open={chatOpen}
        onClose={() => setChatOpen(false)}
        sx={{
          '& .MuiDrawer-paper': {
            width: isMobile ? '100%' : 400,
            height: isMobile ? '80vh' : '100%',
          },
        }}
      >
        <Box sx={{ display: 'flex', justifyContent: 'flex-end', p: 1 }}>
          <IconButton onClick={() => setChatOpen(false)}>
            <CloseIcon />
          </IconButton>
        </Box>
        <Box sx={{ height: 'calc(100% - 56px)', px: 1 }}>
          <ChatPanel />
        </Box>
      </Drawer>
    </Box>
  );
};

export default CalendarPage;
