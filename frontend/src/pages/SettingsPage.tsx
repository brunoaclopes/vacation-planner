import { useState, useEffect } from 'react';
import {
  Box,
  Paper,
  Typography,
  TextField,
  Button,
  Snackbar,
  Alert,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Divider,
  IconButton,
  InputAdornment,
  Grid,
  CircularProgress,
  Autocomplete,
  useTheme,
  alpha,
} from '@mui/material';
import {
  Save as SaveIcon,
  Visibility as VisibilityIcon,
  VisibilityOff as VisibilityOffIcon,
  Refresh as RefreshIcon,
  SmartToy as SmartToyIcon,
  CalendarMonth as CalendarMonthIcon,
  LocationCity as LocationCityIcon,
  MenuBook as MenuBookIcon,
  Language as LanguageIcon,
} from '@mui/icons-material';
import * as api from '../services/api';
import { Settings, WORK_WEEK_PRESETS } from '../types';
import type { AIModel } from '../services/api';
import { useTranslations, useI18n, Language } from '../i18n';

const SettingsPage: React.FC = () => {
  const t = useTranslations();
  const { language, setLanguage } = useI18n();
  const [settings, setSettings] = useState<Settings>({
    openai_api_key: '',
    ai_provider: 'github',
    ai_model: 'gpt-4o-mini',
    default_work_week: JSON.stringify(['monday', 'tuesday', 'wednesday', 'thursday', 'friday']),
    default_vacation_days: '22',
    default_optimization_strategy: 'balanced',
    work_city: '',
    calendarific_api_key: '',
  });
  const [showApiKey, setShowApiKey] = useState(false);
  const [showCalendarificKey, setShowCalendarificKey] = useState(false);
  const [saved, setSaved] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [models, setModels] = useState<AIModel[]>([]);
  const [loadingModels, setLoadingModels] = useState(false);
  const [cities, setCities] = useState<string[]>([]);
  const theme = useTheme();

  useEffect(() => {
    loadSettings();
    loadCities();
  }, []);

  useEffect(() => {
    // Load models when API key is present
    if (settings.openai_api_key) {
      loadModels();
    }
  }, [settings.openai_api_key, settings.ai_provider]);

  const loadSettings = async () => {
    try {
      const data = await api.getSettings();
      setSettings(data);
    } catch (err) {
      setError('Failed to load settings');
    } finally {
      setLoading(false);
    }
  };

  const loadCities = async () => {
    try {
      const availableCities = await api.getAvailableCities();
      setCities(availableCities.sort());
    } catch (err) {
      console.error('Failed to load cities:', err);
    }
  };

  const loadModels = async () => {
    setLoadingModels(true);
    try {
      const availableModels = await api.getAvailableModels();
      setModels(availableModels);
    } catch (err) {
      console.error('Failed to load models:', err);
      // Set default models if fetch fails
      setModels([
        { id: 'gpt-4o-mini', name: 'GPT-4o Mini', publisher: 'openai' },
        { id: 'gpt-4o', name: 'GPT-4o', publisher: 'openai' },
      ]);
    } finally {
      setLoadingModels(false);
    }
  };

  const handleSave = async () => {
    try {
      await api.updateSettings(settings);
      setSaved(true);
      setError(null);
      setTimeout(() => setSaved(false), 3000);
    } catch (err) {
      setError('Failed to save settings');
    }
  };

  const handleChange = (key: keyof Settings, value: string) => {
    setSettings(prev => ({ ...prev, [key]: value }));
  };

  if (loading) {
    return <Typography>{t.common.loading}</Typography>;
  }

  return (
    <Box>
      <Typography 
        variant="h4" 
        gutterBottom
        sx={{
          fontWeight: 700,
          color: 'text.primary',
          mb: 3,
        }}
      >
        {t.settings.title}
      </Typography>

      <Grid container spacing={3}>
        {/* AI Settings */}
        <Grid item xs={12} md={6}>
          <Paper 
            elevation={0} 
            sx={{ 
              p: 3,
              border: '1px solid',
              borderColor: 'grey.200',
              borderRadius: 3,
              height: '100%',
            }}
          >
            <Typography 
              variant="h6" 
              gutterBottom
              sx={{ fontWeight: 600, color: 'text.primary', display: 'flex', alignItems: 'center', gap: 1 }}
            >
              <SmartToyIcon fontSize="small" /> {t.settings.aiIntegration}
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 2.5 }}>
              {t.settings.aiIntegrationDesc}
            </Typography>

            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2.5 }}>
              <FormControl fullWidth size="small">
                <InputLabel>{t.settings.aiProvider}</InputLabel>
                <Select
                  value={settings.ai_provider || 'github'}
                  label={t.settings.aiProvider}
                  onChange={(e) => handleChange('ai_provider', e.target.value)}
                  sx={{ borderRadius: 2 }}
                >
                  <MenuItem value="github">{t.settings.providerGitHub}</MenuItem>
                  <MenuItem value="openai">{t.settings.providerOpenAI}</MenuItem>
                </Select>
              </FormControl>

              <TextField
                fullWidth
                label={settings.ai_provider === 'github' ? t.settings.githubToken : t.settings.openaiKey}
                type={showApiKey ? 'text' : 'password'}
                value={settings.openai_api_key}
                onChange={(e) => handleChange('openai_api_key', e.target.value)}
                placeholder={settings.ai_provider === 'github' ? t.settings.githubTokenPlaceholder : t.settings.openaiKeyPlaceholder}
                InputProps={{
                  endAdornment: (
                    <InputAdornment position="end">
                      <IconButton
                        onClick={() => setShowApiKey(!showApiKey)}
                        edge="end"
                      >
                        {showApiKey ? <VisibilityOffIcon /> : <VisibilityIcon />}
                      </IconButton>
                    </InputAdornment>
                  ),
                }}
                helperText={
                  settings.ai_provider === 'github'
                    ? t.settings.githubTokenHelp
                    : t.settings.openaiKeyHelp
                }
                sx={{
                  '& .MuiOutlinedInput-root': { borderRadius: 2 },
                }}
              />

              <FormControl fullWidth size="small">
                <InputLabel>{t.settings.aiModel}</InputLabel>
                <Select
                  value={settings.ai_model || 'gpt-4o-mini'}
                  label={t.settings.aiModel}
                  onChange={(e) => handleChange('ai_model', e.target.value)}
                  disabled={loadingModels}
                  sx={{ borderRadius: 2 }}
                  endAdornment={
                    loadingModels ? (
                      <CircularProgress size={20} sx={{ mr: 3 }} />
                    ) : (
                      <IconButton
                        size="small"
                        onClick={(e) => {
                          e.stopPropagation();
                          loadModels();
                        }}
                        sx={{ mr: 2 }}
                        title={t.settings.refreshModels}
                      >
                        <RefreshIcon fontSize="small" />
                      </IconButton>
                    )
                  }
                >
                  {models.length > 0 ? (
                    models.map((model) => (
                      <MenuItem key={model.id} value={model.id}>
                        {model.name} ({model.publisher})
                      </MenuItem>
                    ))
                  ) : (
                    <>
                      <MenuItem value="gpt-4o-mini">GPT-4o Mini (openai)</MenuItem>
                      <MenuItem value="gpt-4o">GPT-4o (openai)</MenuItem>
                      <MenuItem value="o3-mini">O3 Mini (openai)</MenuItem>
                      <MenuItem value="Phi-4">Phi-4 (microsoft)</MenuItem>
                      <MenuItem value="Mistral-small">Mistral Small (mistral)</MenuItem>
                    </>
                  )}
                </Select>
              </FormControl>
            </Box>
          </Paper>
        </Grid>

        {/* Server Settings */}
        {/* Language Settings */}
        <Grid item xs={12} md={6}>
          <Paper 
            elevation={0} 
            sx={{ 
              p: 3,
              border: '1px solid',
              borderColor: 'grey.200',
              borderRadius: 3,
              height: '100%',
            }}
          >
            <Typography 
              variant="h6" 
              gutterBottom
              sx={{ fontWeight: 600, color: 'text.primary', display: 'flex', alignItems: 'center', gap: 1 }}
            >
              <LanguageIcon fontSize="small" /> {t.settings.language}
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 2.5 }}>
              {t.settings.languageDesc}
            </Typography>

            <FormControl fullWidth size="small">
              <InputLabel>{t.settings.language}</InputLabel>
              <Select
                value={language}
                label={t.settings.language}
                onChange={(e) => setLanguage(e.target.value as Language)}
                sx={{ borderRadius: 2 }}
              >
                <MenuItem value="en">English</MenuItem>
                <MenuItem value="pt-PT">Português (Portugal)</MenuItem>
              </Select>
            </FormControl>
          </Paper>
        </Grid>

        {/* Default Settings */}
        <Grid item xs={12}>
          <Paper 
            elevation={0} 
            sx={{ 
              p: 3,
              border: '1px solid',
              borderColor: 'grey.200',
              borderRadius: 3,
            }}
          >
            <Typography 
              variant="h6" 
              gutterBottom
              sx={{ fontWeight: 600, color: 'text.primary', display: 'flex', alignItems: 'center', gap: 1 }}
            >
              <CalendarMonthIcon fontSize="small" /> {t.settings.defaultYearConfig}
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 2.5 }}>
              {t.settings.defaultYearConfigDesc}
            </Typography>

            <Grid container spacing={2.5}>
              <Grid item xs={12} sm={4}>
                <TextField
                  fullWidth
                  label={t.settings.defaultVacationDays}
                  type="number"
                  value={settings.default_vacation_days}
                  onChange={(e) => handleChange('default_vacation_days', e.target.value)}
                  inputProps={{ min: 0, max: 365 }}
                  size="small"
                  sx={{
                    '& .MuiOutlinedInput-root': { borderRadius: 2 },
                  }}
                />
              </Grid>

              <Grid item xs={12} sm={4}>
                <FormControl fullWidth size="small">
                  <InputLabel>{t.settings.defaultStrategy}</InputLabel>
                  <Select
                    value={settings.default_optimization_strategy}
                    label={t.settings.defaultStrategy}
                    onChange={(e) => handleChange('default_optimization_strategy', e.target.value)}
                    sx={{ borderRadius: 2 }}
                  >
                    <MenuItem value="bridge_holidays">{t.settings.strategyBridgeHolidays}</MenuItem>
                    <MenuItem value="longest_blocks">{t.settings.strategyLongestBlocks}</MenuItem>
                    <MenuItem value="balanced">{t.settings.strategyBalanced}</MenuItem>
                  </Select>
                </FormControl>
              </Grid>

              <Grid item xs={12} sm={4}>
                <FormControl fullWidth size="small">
                  <InputLabel>{t.settings.defaultWorkWeek}</InputLabel>
                  <Select
                    value={
                      Object.entries(WORK_WEEK_PRESETS).find(
                        ([, days]) => JSON.stringify(days) === settings.default_work_week
                      )?.[0] || 'standard'
                    }
                    label={t.settings.defaultWorkWeek}
                    onChange={(e) => {
                      const preset = e.target.value as string;
                      const days = WORK_WEEK_PRESETS[preset] || WORK_WEEK_PRESETS.standard;
                      handleChange('default_work_week', JSON.stringify(days));
                    }}
                    sx={{ borderRadius: 2 }}
                  >
                    <MenuItem value="standard">{t.config.presetStandard}</MenuItem>
                    <MenuItem value="four_day">{t.config.preset4DayMonThu}</MenuItem>
                    <MenuItem value="four_day_fri">{t.config.preset4DayTueFri}</MenuItem>
                    <MenuItem value="six_day">{t.config.preset6Day}</MenuItem>
                  </Select>
                </FormControl>
              </Grid>
            </Grid>
          </Paper>
        </Grid>

        {/* Work City for Municipal Holidays */}
        <Grid item xs={12} md={6}>
          <Paper 
            elevation={0} 
            sx={{ 
              p: 3,
              border: '1px solid',
              borderColor: 'grey.200',
              borderRadius: 3,
              height: '100%',
            }}
          >
            <Typography 
              variant="h6" 
              gutterBottom
              sx={{ fontWeight: 600, color: 'text.primary', display: 'flex', alignItems: 'center', gap: 1 }}
            >
              <LocationCityIcon fontSize="small" /> {t.settings.locationSettings}
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ mb: 2.5 }}>
              {t.settings.locationSettingsDesc}
            </Typography>

            <Box sx={{ display: 'flex', flexDirection: 'column', gap: 2.5 }}>
              <Autocomplete
                options={cities}
                value={settings.work_city || ''}
                onChange={(_, value) => handleChange('work_city', value || '')}
                renderInput={(params) => (
                  <TextField
                    {...params}
                    label={t.settings.workCity}
                    placeholder={t.settings.workCityPlaceholder}
                    size="small"
                    helperText={t.settings.workCityHelp}
                    sx={{
                      '& .MuiOutlinedInput-root': { borderRadius: 2 },
                    }}
                  />
                )}
                freeSolo
                selectOnFocus
                clearOnBlur
                handleHomeEndKeys
              />

              <TextField
                fullWidth
                label={t.settings.calendarificKey}
                type={showCalendarificKey ? 'text' : 'password'}
                value={settings.calendarific_api_key || ''}
                onChange={(e) => handleChange('calendarific_api_key', e.target.value)}
                placeholder={t.settings.calendarificKeyPlaceholder}
                size="small"
                InputProps={{
                  endAdornment: (
                    <InputAdornment position="end">
                      <IconButton
                        onClick={() => setShowCalendarificKey(!showCalendarificKey)}
                        edge="end"
                      >
                        {showCalendarificKey ? <VisibilityOffIcon /> : <VisibilityIcon />}
                      </IconButton>
                    </InputAdornment>
                  ),
                }}
                helperText={
                  <Box component="span" sx={{ '& a': { color: 'info.main', textDecoration: 'none', '&:hover': { textDecoration: 'underline' } } }}>
                    {t.settings.calendarificKeyHelp}{' '}
                    <a 
                      href="https://calendarific.com/signup" 
                      target="_blank" 
                      rel="noopener noreferrer"
                    >
                      calendarific.com
                    </a>
                  </Box>
                }
                sx={{
                  '& .MuiOutlinedInput-root': { borderRadius: 2 },
                }}
              />
            </Box>
          </Paper>
        </Grid>

        {/* Instructions */}
        <Grid item xs={12}>
          <Paper 
            elevation={0} 
            sx={{ 
              p: 3, 
              backgroundColor: 'background.paper',
              border: '1px solid',
              borderColor: 'divider',
              borderRadius: 3,
            }}
          >
            <Typography 
              variant="h6" 
              gutterBottom
              sx={{ fontWeight: 600, color: 'text.primary', display: 'flex', alignItems: 'center', gap: 1 }}
            >
              <MenuBookIcon fontSize="small" /> {t.settings.howToUse}
            </Typography>
            <Divider sx={{ mb: 2.5 }} />
            
            <Typography variant="subtitle2" gutterBottom sx={{ fontWeight: 600, color: 'text.secondary' }}>
              {t.settings.howToUseStep1}
            </Typography>
            <Typography variant="body2" color="text.secondary" paragraph sx={{ lineHeight: 1.7, '& a': { color: 'info.main', textDecoration: 'none', '&:hover': { textDecoration: 'underline' } } }}>
              {t.settings.howToUseStep1Desc}
            </Typography>

            <Typography variant="subtitle2" gutterBottom sx={{ fontWeight: 600, color: 'text.secondary' }}>
              {t.settings.howToUseStep2}
            </Typography>
            <Typography variant="body2" color="text.secondary" paragraph sx={{ lineHeight: 1.7 }}>
              {t.settings.howToUseStep2Desc}
            </Typography>

            <Typography variant="subtitle2" gutterBottom sx={{ fontWeight: 600, color: 'text.secondary' }}>
              {t.settings.howToUseStep3}
            </Typography>
            <Typography variant="body2" color="text.secondary" paragraph sx={{ lineHeight: 1.7 }}>
              {t.settings.howToUseStep3Desc}
            </Typography>

            <Typography variant="subtitle2" gutterBottom sx={{ fontWeight: 600, color: 'text.secondary' }}>
              {t.settings.howToUseStep4}
            </Typography>
            <Typography variant="body2" color="text.secondary" paragraph sx={{ lineHeight: 1.7 }}>
              {t.settings.howToUseStep4Desc}
            </Typography>

            <Typography variant="subtitle2" gutterBottom sx={{ fontWeight: 600, color: 'text.secondary' }}>
              {t.settings.howToUseStep5}
            </Typography>
            <Typography variant="body2" color="text.secondary" sx={{ lineHeight: 1.7 }}>
              {t.settings.howToUseStep5Desc}
            </Typography>
          </Paper>
        </Grid>
      </Grid>

      <Box sx={{ mt: 4 }}>
        <Button
          variant="contained"
          size="large"
          startIcon={<SaveIcon />}
          onClick={handleSave}
          sx={{
            px: 4,
            py: 1.5,
            fontWeight: 600,
            boxShadow: 'none',
            '&:hover': {
              boxShadow: `0 4px 12px ${alpha(theme.palette.primary.main, 0.3)}`,
            },
          }}
        >
          {t.settings.saveAllSettings}
        </Button>
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
          {t.settings.settingsSaved}
        </Alert>
      </Snackbar>

      <Snackbar
        open={!!error}
        autoHideDuration={5000}
        onClose={() => setError(null)}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}
      >
        <Alert 
          onClose={() => setError(null)} 
          severity="error" 
          variant="filled"
          sx={{ borderRadius: 2 }}
        >
          {error}
        </Alert>
      </Snackbar>
    </Box>
  );
};

export default SettingsPage;
