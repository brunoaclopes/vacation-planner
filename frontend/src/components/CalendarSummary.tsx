import {
  Box,
  Paper,
  Typography,
  LinearProgress,
  Grid,
  Chip,
  alpha,
  useTheme,
} from '@mui/material';
import {
  EventAvailable as VacationIcon,
  Celebration as HolidayIcon,
  Weekend as WeekendIcon,
  TrendingUp as TrendingIcon,
} from '@mui/icons-material';
import { CalendarSummary as CalendarSummaryType } from '../types';
import { useTranslations } from '../i18n';

interface CalendarSummaryProps {
  summary: CalendarSummaryType;
}

interface StatCardProps {
  icon: React.ReactNode;
  value: string | number;
  label: string;
  color: string;
}

const StatCard: React.FC<StatCardProps> = ({ icon, value, label, color }) => (
  <Paper
    elevation={0}
    sx={{
      p: 2.5,
      textAlign: 'center',
      bgcolor: alpha(color, 0.08),
      border: '1px solid',
      borderColor: alpha(color, 0.2),
      borderRadius: 3,
      transition: 'all 0.2s ease',
      '&:hover': {
        transform: 'translateY(-2px)',
        boxShadow: `0 8px 16px ${alpha(color, 0.15)}`,
      },
    }}
  >
    <Box sx={{ color, mb: 1 }}>{icon}</Box>
    <Typography variant="h4" fontWeight={700} sx={{ color, mb: 0.5 }}>
      {value}
    </Typography>
    <Typography variant="body2" color="text.secondary" fontWeight={500}>
      {label}
    </Typography>
  </Paper>
);

const CalendarSummary: React.FC<CalendarSummaryProps> = ({ summary }) => {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const t = useTranslations();
  const usagePercentage = (summary.used_vacation_days / summary.total_vacation_days) * 100;

  // Theme-aware colors from palette
  const colors = {
    vacation: theme.palette.primary.main,
    holiday: theme.palette.secondary.main,
    success: theme.palette.success.main,
    info: theme.palette.info.main,
    optimized: (theme.palette as any).optimized?.main || theme.palette.info.main,
  };

  return (
    <Paper 
      elevation={0} 
      sx={{ 
        p: 3, 
        mb: 3,
        border: '1px solid',
        borderColor: 'divider',
        borderRadius: 4,
      }}
    >
      <Typography variant="h6" fontWeight={700} gutterBottom sx={{ mb: 3 }}>
        {t.summary.title}
      </Typography>
      
      <Grid container spacing={2}>
        <Grid item xs={6} md={3}>
          <StatCard
            icon={<VacationIcon sx={{ fontSize: 36 }} />}
            value={`${summary.used_vacation_days}/${summary.total_vacation_days}`}
            label={t.summary.vacationDaysUsed}
            color={colors.vacation}
          />
        </Grid>
        
        <Grid item xs={6} md={3}>
          <StatCard
            icon={<HolidayIcon sx={{ fontSize: 36 }} />}
            value={summary.total_holidays}
            label={t.summary.holidaysThisYear}
            color={colors.holiday}
          />
        </Grid>
        
        <Grid item xs={6} md={3}>
          <StatCard
            icon={<TrendingIcon sx={{ fontSize: 36 }} />}
            value={summary.longest_vacation_block}
            label={t.summary.longestBlock}
            color={colors.success}
          />
        </Grid>
        
        <Grid item xs={6} md={3}>
          <StatCard
            icon={<WeekendIcon sx={{ fontSize: 36 }} />}
            value={summary.total_days_off}
            label={t.summary.totalDaysOff}
            color={colors.info}
          />
        </Grid>
      </Grid>

      <Box sx={{ mt: 3 }}>
        <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
          <Typography variant="body2" fontWeight={500}>{t.summary.vacationDaysUsage}</Typography>
          <Typography variant="body2" fontWeight={600} color="primary">
            {summary.remaining_vacation_days} {t.common.days} {t.common.remaining}
          </Typography>
        </Box>
        <LinearProgress
          variant="determinate"
          value={usagePercentage}
          sx={{ 
            height: 10, 
            borderRadius: 5,
            bgcolor: isDark ? 'grey.800' : 'grey.100',
            '& .MuiLinearProgress-bar': {
              borderRadius: 5,
            },
          }}
        />
      </Box>

      <Box sx={{ mt: 3, display: 'flex', gap: 1, flexWrap: 'wrap' }}>
        <Chip
          icon={<HolidayIcon />}
          label={t.calendar.holiday}
          size="small"
          sx={{ 
            bgcolor: alpha(colors.holiday, 0.1), 
            color: colors.holiday,
            fontWeight: 500,
            '& .MuiChip-icon': { color: colors.holiday },
          }}
        />
        <Chip
          icon={<VacationIcon />}
          label={t.calendar.manualVacation}
          size="small"
          sx={{ 
            bgcolor: alpha(colors.vacation, 0.1), 
            color: colors.vacation,
            fontWeight: 500,
            '& .MuiChip-icon': { color: colors.vacation },
          }}
        />
        <Chip
          label={t.calendar.optimizedVacation}
          size="small"
          sx={{ 
            bgcolor: alpha(colors.optimized, 0.1), 
            color: colors.optimized,
            fontWeight: 500,
          }}
        />
        <Chip
          icon={<WeekendIcon />}
          label={t.calendar.weekend}
          size="small"
          variant="outlined"
          sx={{ 
            fontWeight: 500,
            borderColor: isDark ? 'grey.600' : 'grey.400',
            color: isDark ? 'grey.400' : 'grey.600',
            '& .MuiChip-icon': { color: isDark ? 'grey.400' : 'grey.600' },
          }}
        />
      </Box>
    </Paper>
  );
};

export default CalendarSummary;
