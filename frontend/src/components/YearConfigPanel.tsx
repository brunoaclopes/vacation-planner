import { useState, useEffect } from 'react';
import {
  Box,
  Paper,
  Typography,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  TextField,
  Chip,
  FormGroup,
  FormControlLabel,
  Checkbox,
  Button,
  Divider,
  Snackbar,
  Alert,
  SelectChangeEvent,
  useTheme,
  alpha,
} from '@mui/material';
import { Save as SaveIcon, DeleteSweep as ClearIcon } from '@mui/icons-material';
import { useCalendar } from '../context/CalendarContext';
import { OptimizationStrategy, ALL_WEEKDAYS, WORK_WEEK_PRESETS } from '../types';
import * as api from '../services/api';
import { useTranslations, interpolate } from '../i18n';

const YearConfigPanel: React.FC = () => {
  const { calendar, updateConfig, clearOptimized, year } = useCalendar();
  const t = useTranslations();
  const [strategies, setStrategies] = useState<OptimizationStrategy[]>([]);
  const [vacationDays, setVacationDays] = useState<number>(22);
  const [reservedDays, setReservedDays] = useState<number>(0);
  const [strategy, setStrategy] = useState<string>('balanced');
  const [workWeek, setWorkWeek] = useState<string[]>(['monday', 'tuesday', 'wednesday', 'thursday', 'friday']);
  const [workWeekPreset, setWorkWeekPreset] = useState<string>('standard');
  const [optimizerNotes, setOptimizerNotes] = useState<string>('');
  const [saved, setSaved] = useState(false);

  useEffect(() => {
    api.getOptimizationStrategies().then(setStrategies).catch(console.error);
  }, []);

  useEffect(() => {
    if (calendar?.config) {
      setVacationDays(calendar.config.vacation_days);
      setReservedDays(calendar.config.reserved_days || 0);
      setStrategy(calendar.config.optimization_strategy);
      setWorkWeek(calendar.config.work_week);
      setOptimizerNotes(calendar.config.optimizer_notes || '');
      
      // Determine preset
      const preset = Object.entries(WORK_WEEK_PRESETS).find(
        ([, days]) => JSON.stringify(days.sort()) === JSON.stringify([...calendar.config.work_week].sort())
      );
      setWorkWeekPreset(preset ? preset[0] : 'custom');
    }
  }, [calendar]);

  const handleWorkWeekPresetChange = (e: SelectChangeEvent<string>) => {
    const preset = e.target.value;
    setWorkWeekPreset(preset);
    if (preset !== 'custom' && WORK_WEEK_PRESETS[preset]) {
      setWorkWeek(WORK_WEEK_PRESETS[preset]);
    }
  };

  const handleWorkDayToggle = (day: string) => {
    setWorkWeekPreset('custom');
    if (workWeek.includes(day)) {
      setWorkWeek(workWeek.filter(d => d !== day));
    } else {
      setWorkWeek([...workWeek, day]);
    }
  };

  const handleSave = async () => {
    await updateConfig({
      vacation_days: vacationDays,
      reserved_days: reservedDays,
      optimization_strategy: strategy,
      work_week: workWeek,
      optimizer_notes: optimizerNotes,
    });
    setSaved(true);
    setTimeout(() => setSaved(false), 3000);
  };

  const capitalizeFirst = (str: string) => str.charAt(0).toUpperCase() + str.slice(1);
  const theme = useTheme();

  return (
    <Paper 
      elevation={0} 
      sx={{ 
        p: 3, 
        mb: 2,
        border: '1px solid',
        borderColor: 'divider',
        borderRadius: 3,
      }}
    >
      <Typography 
        variant="h6" 
        gutterBottom
        sx={{ 
          fontWeight: 600,
          color: 'text.primary',
          mb: 2.5,
        }}
      >
        {t.config.title} ({year})
      </Typography>

      <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2.5 }}>
        <TextField
          label={t.config.totalVacationDays}
          type="number"
          value={vacationDays}
          onChange={(e) => setVacationDays(parseInt(e.target.value) || 0)}
          inputProps={{ min: 0, max: 365 }}
          helperText={t.config.totalVacationDaysHelp}
          size="small"
          sx={{
            '& .MuiOutlinedInput-root': {
              borderRadius: 2,
            },
          }}
        />

        <TextField
          label={t.config.reservedDays}
          type="number"
          value={reservedDays}
          onChange={(e) => setReservedDays(parseInt(e.target.value) || 0)}
          inputProps={{ min: 0, max: vacationDays }}
          helperText={interpolate(t.config.reservedDaysHelp, [reservedDays, vacationDays - reservedDays])}
          size="small"
          sx={{
            '& .MuiOutlinedInput-root': {
              borderRadius: 2,
            },
          }}
        />

        <FormControl size="small">
          <InputLabel>{t.config.optimizationStrategy}</InputLabel>
          <Select
            value={strategy}
            label={t.config.optimizationStrategy}
            onChange={(e) => setStrategy(e.target.value)}
            renderValue={(selected) => {
              const selectedStrategy = strategies.find(s => s.id === selected);
              return selectedStrategy?.name || selected;
            }}
            sx={{
              borderRadius: 2,
            }}
          >
            {strategies.map((s) => (
              <MenuItem key={s.id} value={s.id}>
                <Box>
                  <Typography sx={{ fontWeight: 500 }}>{s.name}</Typography>
                  <Typography variant="caption" color="text.secondary">
                    {s.description}
                  </Typography>
                </Box>
              </MenuItem>
            ))}
          </Select>
        </FormControl>

        {strategy === 'smart' && (
          <TextField
            label={t.config.smartOptimizerNotes}
            value={optimizerNotes}
            onChange={(e) => setOptimizerNotes(e.target.value)}
            multiline
            rows={3}
            placeholder={t.config.smartOptimizerNotesPlaceholder}
            helperText={t.config.smartOptimizerNotesHelp}
            size="small"
            sx={{
              '& .MuiOutlinedInput-root': {
                borderRadius: 2,
              },
            }}
          />
        )}

        <Divider sx={{ my: 0.5 }} />

        <Typography 
          variant="subtitle2"
          sx={{ 
            fontWeight: 600,
            color: 'text.secondary',
          }}
        >
          {t.config.workWeekConfig}
        </Typography>
        
        <FormControl size="small">
          <InputLabel>{t.config.preset}</InputLabel>
          <Select
            value={workWeekPreset}
            label={t.config.preset}
            onChange={handleWorkWeekPresetChange}
            sx={{
              borderRadius: 2,
            }}
          >
            <MenuItem value="standard">{t.config.presetStandard}</MenuItem>
            <MenuItem value="four_day">{t.config.preset4DayMonThu}</MenuItem>
            <MenuItem value="four_day_fri">{t.config.preset4DayTueFri}</MenuItem>
            <MenuItem value="six_day">{t.config.preset6Day}</MenuItem>
            <MenuItem value="custom">{t.config.presetCustom}</MenuItem>
          </Select>
        </FormControl>

        <FormGroup row sx={{ gap: 0.5 }}>
          {ALL_WEEKDAYS.map((day) => (
            <FormControlLabel
              key={day}
              control={
                <Checkbox
                  checked={workWeek.includes(day)}
                  onChange={() => handleWorkDayToggle(day)}
                  size="small"
                  sx={{
                    '&.Mui-checked': {
                      color: 'primary.main',
                    },
                  }}
                />
              }
              label={capitalizeFirst(day.slice(0, 3))}
              sx={{
                '& .MuiFormControlLabel-label': {
                  fontSize: '0.875rem',
                  fontWeight: 500,
                },
              }}
            />
          ))}
        </FormGroup>

        <Box sx={{ display: 'flex', gap: 1, flexWrap: 'wrap', alignItems: 'center' }}>
          <Typography variant="body2" color="text.secondary" sx={{ fontWeight: 500 }}>
            Work days:
          </Typography>
          {workWeek.sort((a, b) => ALL_WEEKDAYS.indexOf(a) - ALL_WEEKDAYS.indexOf(b)).map((day) => (
            <Chip 
              key={day} 
              label={capitalizeFirst(day)} 
              size="small" 
              color="primary"
              sx={{
                fontWeight: 500,
                borderRadius: 1.5,
              }}
            />
          ))}
        </Box>

        <Box sx={{ display: 'flex', gap: 2, flexWrap: 'wrap', mt: 1 }}>
          <Button
            variant="contained"
            startIcon={<SaveIcon />}
            onClick={handleSave}
            sx={{
              px: 3,
              py: 1,
              fontWeight: 600,
              boxShadow: 'none',
              '&:hover': {
                boxShadow: `0 4px 12px ${alpha(theme.palette.primary.main, 0.3)}`,
              },
            }}
          >
            {t.config.saveConfig}
          </Button>
          <Button
            variant="outlined"
            color="warning"
            startIcon={<ClearIcon />}
            onClick={clearOptimized}
            sx={{
              px: 3,
              py: 1,
              fontWeight: 600,
              borderWidth: 2,
              '&:hover': {
                borderWidth: 2,
              },
            }}
          >
            {t.calendar.clearOptimized}
          </Button>
        </Box>
      </Box>

      <Snackbar
        open={saved}
        autoHideDuration={3000}
        onClose={() => setSaved(false)}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert 
          onClose={() => setSaved(false)} 
          severity="success" 
          variant="filled"
          sx={{ borderRadius: 2 }}
        >
          {t.config.configSaved}
        </Alert>
      </Snackbar>
    </Paper>
  );
};

export default YearConfigPanel;
