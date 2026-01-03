import React, { useState, useEffect, useCallback } from 'react';
import {
  Snackbar,
  Alert,
  AlertTitle,
  IconButton,
  Box,
  Typography,
  LinearProgress,
  Collapse,
  Button,
} from '@mui/material';
import {
  Close as CloseIcon,
  Refresh as RefreshIcon,
  ExpandMore as ExpandMoreIcon,
  ExpandLess as ExpandLessIcon,
} from '@mui/icons-material';
import * as api from '../services/api';
import { useTranslations, interpolate } from '../i18n';

interface HolidayNotificationProps {
  year: number;
  onRefresh?: () => void;
}

const HolidayNotification: React.FC<HolidayNotificationProps> = ({ year, onRefresh }) => {
  const t = useTranslations();
  const [status, setStatus] = useState<api.HolidayStatus | null>(null);
  const [open, setOpen] = useState(false);
  const [expanded, setExpanded] = useState(false);
  const [dismissed, setDismissed] = useState(false);
  const [refreshing, setRefreshing] = useState(false);

  const checkStatus = useCallback(async () => {
    if (dismissed) return;
    
    try {
      const statusData = await api.getHolidayStatus(year);
      setStatus(statusData);
      
      // Show notification if there are errors
      if (statusData.has_errors) {
        setOpen(true);
      } else {
        setOpen(false);
      }
    } catch (error) {
      // Silently fail - endpoint might not exist yet
      console.debug('Holiday status check failed:', error);
    }
  }, [year, dismissed]);

  useEffect(() => {
    // Initial check
    checkStatus();

    // Poll every 10 seconds if there are errors and retrying
    const interval = setInterval(() => {
      if (status?.is_retrying) {
        checkStatus();
      }
    }, 10000);

    return () => clearInterval(interval);
  }, [checkStatus, status?.is_retrying]);

  // Re-check when year changes
  useEffect(() => {
    setDismissed(false);
    checkStatus();
  }, [year, checkStatus]);

  const handleClose = () => {
    setOpen(false);
    setDismissed(true);
  };

  const handleRefresh = async () => {
    setRefreshing(true);
    try {
      await api.refreshHolidays(year);
      await checkStatus();
      onRefresh?.();
    } catch (error) {
      console.error('Failed to refresh holidays:', error);
    } finally {
      setRefreshing(false);
    }
  };

  if (!status?.has_errors || !open) {
    return null;
  }

  const getErrorMessage = () => {
    const errors: string[] = [];
    if (status.national_error) {
      errors.push(`National holidays: ${status.national_error}`);
    }
    if (status.municipal_error) {
      errors.push(`Municipal holidays: ${status.municipal_error}`);
    }
    return errors;
  };

  const getRetryInfo = () => {
    if (status.is_retrying) {
      return interpolate(t.holidays.retrying, [status.retry_count, status.max_retries]);
    }
    if (status.retry_count >= status.max_retries) {
      return t.holidays.maxRetriesReached;
    }
    return null;
  };

  return (
    <Snackbar
      open={open}
      anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
      sx={{ maxWidth: 500, width: '100%' }}
    >
      <Alert
        severity="warning"
        variant="filled"
        sx={{ width: '100%' }}
        action={
          <Box sx={{ display: 'flex', alignItems: 'flex-start', gap: 0.5 }}>
            <IconButton
              size="small"
              color="inherit"
              onClick={handleRefresh}
              disabled={refreshing}
              title="Refresh holidays"
            >
              <RefreshIcon fontSize="small" />
            </IconButton>
            <IconButton
              size="small"
              color="inherit"
              onClick={() => setExpanded(!expanded)}
              title={expanded ? t.holidays.hideDetails : t.holidays.showDetails}
            >
              {expanded ? <ExpandLessIcon fontSize="small" /> : <ExpandMoreIcon fontSize="small" />}
            </IconButton>
            <IconButton
              size="small"
              color="inherit"
              onClick={handleClose}
              title={t.holidays.dismiss}
            >
              <CloseIcon fontSize="small" />
            </IconButton>
          </Box>
        }
      >
        <AlertTitle sx={{ fontWeight: 600 }}>{t.holidays.errorTitle}</AlertTitle>
        <Typography variant="body2">
          {t.holidays.errorDesc}
        </Typography>
        
        {status.is_retrying && (
          <Box sx={{ mt: 1 }}>
            <LinearProgress color="inherit" sx={{ opacity: 0.7, borderRadius: 1 }} />
          </Box>
        )}
        
        <Collapse in={expanded}>
          <Box sx={{ mt: 2 }}>
            {getErrorMessage().map((error, index) => (
              <Typography key={index} variant="caption" display="block" sx={{ opacity: 0.9 }}>
                • {error}
              </Typography>
            ))}
            
            {getRetryInfo() && (
              <Typography variant="caption" display="block" sx={{ mt: 1, fontStyle: 'italic' }}>
                {getRetryInfo()}
              </Typography>
            )}
            
            {!status.is_retrying && status.retry_count >= status.max_retries && (
              <Button
                size="small"
                color="inherit"
                variant="outlined"
                startIcon={<RefreshIcon />}
                onClick={handleRefresh}
                disabled={refreshing}
                sx={{ mt: 1 }}
              >
                {t.holidays.retryNow}
              </Button>
            )}
          </Box>
        </Collapse>
      </Alert>
    </Snackbar>
  );
};

export default HolidayNotification;
