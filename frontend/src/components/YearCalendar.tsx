import { useMemo } from 'react';
import {
  Box,
  Paper,
  Typography,
  Tooltip,
  useTheme,
  alpha,
} from '@mui/material';
import { format, startOfMonth, endOfMonth, eachDayOfInterval, getDay, isSameMonth } from 'date-fns';
import { CalendarDay } from '../types';
import { useTranslations } from '../i18n';

interface YearCalendarProps {
  year: number;
  days: CalendarDay[];
  onDayClick: (day: CalendarDay) => void;
}

const YearCalendar: React.FC<YearCalendarProps> = ({ year, days, onDayClick }) => {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const t = useTranslations();

  const daysMap = useMemo(() => {
    const map = new Map<string, CalendarDay>();
    days.forEach(day => map.set(day.date, day));
    return map;
  }, [days]);

  // Theme-aware colors from palette
  const colors = {
    holiday: theme.palette.secondary.main,
    manual: theme.palette.primary.main,
    optimized: (theme.palette as any).optimized?.main || theme.palette.info.main,
    weekend: alpha(theme.palette.grey[400], isDark ? 0.4 : 0.2),
    default: isDark ? 'transparent' : theme.palette.background.paper,
  };

  const getDayColor = (day: CalendarDay | undefined) => {
    if (!day) return 'transparent';
    if (day.is_holiday) return colors.holiday;
    if (day.is_manual) return colors.manual;
    if (day.is_optimal) return colors.optimized;
    if (day.is_weekend) return colors.weekend;
    return colors.default;
  };

  const getDayTextColor = (day: CalendarDay | undefined) => {
    if (!day) return 'transparent';
    if (day.is_holiday) return 'white';
    if (day.is_manual) return 'white';
    if (day.is_optimal) return 'white';
    if (day.is_weekend) return alpha(theme.palette.grey[500], isDark ? 0.7 : 1);
    return theme.palette.text.primary;
  };

  const getTooltipContent = (day: CalendarDay | undefined) => {
    if (!day) return '';
    const parts = [format(new Date(day.date), 'EEEE, MMMM d, yyyy')];
    if (day.holiday_name) parts.push(`🎉 ${t.calendar.holiday}: ${day.holiday_name}`);
    if (day.is_manual) parts.push(t.calendar.tooltipManualVacation);
    if (day.is_optimal) parts.push(`${t.calendar.tooltipOptimizedVacation} (Block ${day.block_id})`);
    if (day.is_weekend) parts.push(t.calendar.weekend);
    return parts.join('\n');
  };

  const renderMonth = (monthIndex: number) => {
    const monthStart = startOfMonth(new Date(year, monthIndex));
    const monthEnd = endOfMonth(monthStart);
    const daysInMonth = eachDayOfInterval({ start: monthStart, end: monthEnd });
    const startDayOfWeek = getDay(monthStart);

    const grid: (Date | null)[] = [];
    for (let i = 0; i < startDayOfWeek; i++) {
      grid.push(null);
    }
    daysInMonth.forEach(day => grid.push(day));

    return (
      <Paper
        key={monthIndex}
        elevation={0}
        sx={{
          p: 2,
          minWidth: 220,
          flex: '1 1 calc(25% - 16px)',
          minHeight: 220,
          border: '1px solid',
          borderColor: 'divider',
          borderRadius: 3,
          transition: 'all 0.2s ease',
          '&:hover': {
            borderColor: isDark ? 'grey.600' : 'grey.300',
            boxShadow: `0 4px 12px ${alpha('#000000', isDark ? 0.3 : 0.05)}`,
          },
        }}
      >
        <Typography
          variant="subtitle1"
          fontWeight={700}
          textAlign="center"
          sx={{ mb: 1.5, color: 'text.primary' }}
        >
          {t.calendar.months[monthIndex]}
        </Typography>
        
        <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)', gap: 0.5 }}>
          {t.calendar.weekdaysShort.map((day, i) => (
            <Typography
              key={`${day}-${i}`}
              variant="caption"
              textAlign="center"
              sx={{ 
                fontWeight: 600, 
                color: 'text.disabled', 
                fontSize: '0.7rem',
                mb: 0.5,
              }}
            >
              {day}
            </Typography>
          ))}
        </Box>

        <Box sx={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)', gap: 0.5 }}>
          {grid.map((date, index) => {
            if (!date) {
              return <Box key={`empty-${index}`} sx={{ aspectRatio: '1' }} />;
            }

            const dateStr = format(date, 'yyyy-MM-dd');
            const calendarDay = daysMap.get(dateStr);
            const isCurrentMonth = isSameMonth(date, monthStart);
            const isSpecialDay = calendarDay?.is_holiday || calendarDay?.is_manual || calendarDay?.is_optimal;

            return (
              <Tooltip
                key={dateStr}
                title={<span style={{ whiteSpace: 'pre-line' }}>{getTooltipContent(calendarDay)}</span>}
                arrow
                placement="top"
              >
                <Box
                  onClick={() => calendarDay && onDayClick(calendarDay)}
                  sx={{
                    aspectRatio: '1',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    backgroundColor: getDayColor(calendarDay),
                    color: getDayTextColor(calendarDay),
                    borderRadius: '50%',
                    cursor: 'pointer',
                    fontSize: '0.75rem',
                    fontWeight: isSpecialDay ? 600 : 400,
                    opacity: isCurrentMonth ? 1 : 0.4,
                    transition: 'all 0.15s ease',
                    '&:hover': {
                      transform: 'scale(1.15)',
                      boxShadow: isSpecialDay 
                        ? `0 4px 8px ${alpha(getDayColor(calendarDay) as string, 0.4)}`
                        : `0 2px 4px ${alpha('#000000', 0.1)}`,
                    },
                  }}
                >
                  {format(date, 'd')}
                </Box>
              </Tooltip>
            );
          })}
        </Box>
      </Paper>
    );
  };

  return (
    <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 2 }}>
      {Array.from({ length: 12 }, (_, i) => renderMonth(i))}
    </Box>
  );
};

export default YearCalendar;
