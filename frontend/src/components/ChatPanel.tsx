import { useState, useRef, useEffect } from 'react';
import {
  Box,
  Paper,
  Typography,
  TextField,
  IconButton,
  List,
  ListItem,
  CircularProgress,
  Divider,
  Button,
  useTheme,
  alpha,
} from '@mui/material';
import {
  Send as SendIcon,
  Delete as DeleteIcon,
  SmartToy as BotIcon,
  Person as PersonIcon,
} from '@mui/icons-material';
import { useCalendar } from '../context/CalendarContext';
import { useTranslations } from '../i18n';

const ChatPanel: React.FC = () => {
  const theme = useTheme();
  const isDark = theme.palette.mode === 'dark';
  const t = useTranslations();
  const [message, setMessage] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const { chatMessages, sendMessage, clearChat, chatLoading, loadChatHistory } = useCalendar();

  useEffect(() => {
    loadChatHistory();
  }, [loadChatHistory]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [chatMessages]);

  const handleSend = async () => {
    if (!message.trim() || chatLoading) return;
    const msg = message;
    setMessage('');
    await sendMessage(msg);
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const formatMessageContent = (content: string) => {
    // Remove JSON action blocks from display (both inline and code-fenced)
    let cleanContent = content;
    
    // Remove markdown code blocks containing JSON actions
    cleanContent = cleanContent.replace(/```json\s*\{[^`]*"action"[^`]*\}\s*```/gs, '');
    cleanContent = cleanContent.replace(/```\s*\{[^`]*"action"[^`]*\}\s*```/gs, '');
    
    // Remove inline JSON action blocks (handles nested braces)
    cleanContent = cleanContent.replace(/\{"action"\s*:\s*"[^"]+"\s*,\s*"[^}]+\}/g, '');
    cleanContent = cleanContent.replace(/\{"action"\s*:\s*"[^"]+"\s*\}/g, '');
    
    // Clean up extra whitespace and newlines left behind
    cleanContent = cleanContent.replace(/\n{3,}/g, '\n\n');
    
    return cleanContent.trim();
  };

  return (
    <Paper
      elevation={0}
      sx={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        maxHeight: 'calc(100vh - 200px)',
        border: '1px solid',
        borderColor: 'divider',
        borderRadius: 3,
        overflow: 'hidden',
      }}
    >
      <Box
        sx={{
          p: 2,
          borderBottom: 1,
          borderColor: 'divider',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          backgroundColor: isDark ? 'grey.900' : 'grey.50',
        }}
      >
        <Typography 
          variant="h6"
          sx={{
            fontWeight: 600,
            color: 'text.primary',
            display: 'flex',
            alignItems: 'center',
          }}
        >
          <BotIcon sx={{ mr: 1.5, color: 'primary.main' }} />
          {t.chat.title}
        </Typography>
        <Button
          size="small"
          startIcon={<DeleteIcon />}
          onClick={clearChat}
          color="error"
          sx={{
            fontWeight: 500,
            borderRadius: 2,
          }}
        >
          {t.common.clear}
        </Button>
      </Box>

      <List
        sx={{
          flexGrow: 1,
          overflow: 'auto',
          p: 2,
          backgroundColor: isDark ? 'background.default' : 'grey.50',
        }}
      >
        {chatMessages.length === 0 && (
          <Box sx={{ textAlign: 'center', color: 'text.secondary', mt: 4, px: 2 }}>
            <Box
              sx={{
                width: 64,
                height: 64,
                borderRadius: '50%',
                backgroundColor: 'primary.main',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                mx: 'auto',
                mb: 2,
                opacity: 0.8,
              }}
            >
              <BotIcon sx={{ fontSize: 32, color: isDark ? 'grey.900' : 'white' }} />
            </Box>
            <Typography sx={{ fontWeight: 600, color: 'text.primary' }}>
              {t.chat.emptyState}
            </Typography>
            <Typography variant="body2" sx={{ mt: 1, color: 'text.secondary' }}>
              {t.chat.emptyStateHint}
            </Typography>
          </Box>
        )}
        
        {chatMessages.map((msg, index) => (
          <ListItem
            key={msg.id || index}
            sx={{
              flexDirection: 'column',
              alignItems: msg.role === 'user' ? 'flex-end' : 'flex-start',
              p: 1,
            }}
          >
            <Box
              sx={{
                display: 'flex',
                alignItems: 'flex-start',
                maxWidth: '85%',
                flexDirection: msg.role === 'user' ? 'row-reverse' : 'row',
              }}
            >
              <Box
                sx={{
                  width: 36,
                  height: 36,
                  borderRadius: '50%',
                  backgroundColor: msg.role === 'user' ? 'primary.main' : 'grey.700',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  mx: 1.5,
                  boxShadow: `0 2px 8px ${alpha('#000000', 0.15)}`,
                }}
              >
                {msg.role === 'user' ? (
                  <PersonIcon sx={{ color: 'white', fontSize: 20 }} />
                ) : (
                  <BotIcon sx={{ color: 'white', fontSize: 20 }} />
                )}
              </Box>
              <Paper
                elevation={0}
                sx={{
                  p: 2,
                  backgroundColor: msg.role === 'user' ? 'primary.main' : 'background.paper',
                  color: msg.role === 'user' ? (isDark ? 'grey.900' : 'white') : 'text.primary',
                  borderRadius: 3,
                  maxWidth: '100%',
                  border: msg.role === 'user' ? 'none' : '1px solid',
                  borderColor: 'divider',
                  boxShadow: msg.role === 'user' 
                    ? `0 2px 8px ${alpha(theme.palette.primary.main, 0.25)}` 
                    : `0 1px 3px ${alpha('#000000', isDark ? 0.3 : 0.08)}`,
                }}
              >
                <Typography
                  variant="body2"
                  sx={{ 
                    whiteSpace: 'pre-wrap', 
                    wordBreak: 'break-word',
                    lineHeight: 1.6,
                  }}
                >
                  {formatMessageContent(msg.content)}
                </Typography>
              </Paper>
            </Box>
          </ListItem>
        ))}
        
        {chatLoading && (
          <ListItem sx={{ justifyContent: 'flex-start', p: 1 }}>
            <Box sx={{ display: 'flex', alignItems: 'center' }}>
              <Box
                sx={{
                  width: 36,
                  height: 36,
                  borderRadius: '50%',
                  backgroundColor: 'grey.700',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  mx: 1.5,
                  boxShadow: `0 2px 8px ${alpha('#000000', 0.15)}`,
                }}
              >
                <BotIcon sx={{ color: 'white', fontSize: 20 }} />
              </Box>
              <CircularProgress size={24} thickness={4} />
            </Box>
          </ListItem>
        )}
        
        <div ref={messagesEndRef} />
      </List>

      <Divider />
      
      <Box sx={{ p: 2.5, display: 'flex', gap: 1.5, backgroundColor: 'background.paper' }}>
        <TextField
          fullWidth
          size="small"
          placeholder={t.chat.placeholder}
          value={message}
          onChange={(e) => setMessage(e.target.value)}
          onKeyPress={handleKeyPress}
          disabled={chatLoading}
          multiline
          maxRows={3}
          sx={{
            '& .MuiOutlinedInput-root': {
              borderRadius: 2.5,
              backgroundColor: isDark ? 'grey.100' : 'grey.50',
            },
          }}
        />
        <IconButton
          color="primary"
          onClick={handleSend}
          disabled={!message.trim() || chatLoading}
          sx={{
            backgroundColor: 'primary.main',
            color: isDark ? 'grey.900' : 'white',
            width: 44,
            height: 44,
            '&:hover': {
              backgroundColor: 'primary.dark',
            },
            '&:disabled': {
              backgroundColor: isDark ? 'grey.700' : 'grey.300',
              color: isDark ? 'grey.500' : 'grey.500',
            },
          }}
        >
          <SendIcon />
        </IconButton>
      </Box>
    </Paper>
  );
};

export default ChatPanel;
