import { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import {
  AppBar,
  Box,
  Drawer,
  IconButton,
  List,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Toolbar,
  Typography,
  useTheme,
  useMediaQuery,
  Avatar,
  Divider,
  Menu,
  MenuItem,
} from '@mui/material';
import {
  Menu as MenuIcon,
  CalendarMonth as CalendarIcon,
  Settings as SettingsIcon,
  BeachAccess as LogoIcon,
  LightMode as LightModeIcon,
  DarkMode as DarkModeIcon,
  SettingsBrightness as SystemModeIcon,
  ChevronLeft as ChevronLeftIcon,
} from '@mui/icons-material';
import { useThemeMode } from '../context/ThemeContext';
import { useTranslations } from '../i18n';

const drawerWidth = 260;
const collapsedDrawerWidth = 72;

interface LayoutProps {
  children: React.ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const [mobileOpen, setMobileOpen] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [themeMenuAnchor, setThemeMenuAnchor] = useState<null | HTMLElement>(null);
  const navigate = useNavigate();
  const location = useLocation();
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('md'));
  const { mode, setMode, resolvedMode } = useThemeMode();
  const isDark = theme.palette.mode === 'dark';
  const t = useTranslations();

  const currentDrawerWidth = sidebarCollapsed ? collapsedDrawerWidth : drawerWidth;

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen);
  };

  const handleSidebarCollapse = () => {
    setSidebarCollapsed(!sidebarCollapsed);
  };

  const handleThemeMenuOpen = (event: React.MouseEvent<HTMLElement>) => {
    setThemeMenuAnchor(event.currentTarget);
  };

  const handleThemeMenuClose = () => {
    setThemeMenuAnchor(null);
  };

  const handleThemeChange = (newMode: 'light' | 'dark' | 'system') => {
    setMode(newMode);
    handleThemeMenuClose();
  };

  const getThemeIcon = () => {
    if (mode === 'system') return <SystemModeIcon />;
    if (resolvedMode === 'dark') return <DarkModeIcon />;
    return <LightModeIcon />;
  };

  const menuItems = [
    { text: t.menu.calendar, icon: <CalendarIcon />, path: '/' },
    { text: t.menu.settings, icon: <SettingsIcon />, path: '/settings' },
  ];

  const drawer = (
    <Box 
      onClick={sidebarCollapsed ? handleSidebarCollapse : undefined}
      sx={{ 
        height: '100%', 
        display: 'flex', 
        flexDirection: 'column',
        cursor: sidebarCollapsed ? 'pointer' : 'default',
      }}
    >
      <Box 
        sx={{ 
          p: sidebarCollapsed ? 1.5 : 2, 
          pl: sidebarCollapsed ? 1.5 : 3,
          display: 'flex', 
          alignItems: 'center', 
          justifyContent: sidebarCollapsed ? 'center' : 'space-between',
        }}
      >
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
          <Avatar
            sx={{
              bgcolor: 'primary.main',
              width: 44,
              height: 44,
            }}
          >
            <LogoIcon />
          </Avatar>
          {!sidebarCollapsed && (
            <Box>
              <Typography variant="subtitle1" fontWeight={700} color="text.primary">
                {t.app.title}
              </Typography>
              <Typography variant="caption" color="text.secondary">
                {t.app.subtitle}
              </Typography>
            </Box>
          )}
        </Box>
        {/* Collapse toggle button - only show when expanded */}
        {!sidebarCollapsed && (
          <IconButton
            onClick={handleSidebarCollapse}
            size="small"
            sx={{
              display: { xs: 'none', md: 'flex' },
              color: 'text.secondary',
              '&:hover': {
                bgcolor: isDark ? 'grey.200' : 'grey.100',
              },
            }}
          >
            <ChevronLeftIcon />
          </IconButton>
        )}
      </Box>
      <Divider sx={{ mx: sidebarCollapsed ? 1 : 2 }} />
      <List sx={{ px: sidebarCollapsed ? 1 : 2, py: 2, flex: 1 }}>
        {menuItems.map((item) => (
          <ListItem key={item.text} disablePadding sx={{ mb: 0.5 }}>
            <ListItemButton
              selected={location.pathname === item.path}
              onClick={(e) => {
                if (sidebarCollapsed) {
                  e.stopPropagation(); // Don't trigger sidebar expand
                }
                navigate(item.path);
                if (isMobile) setMobileOpen(false);
              }}
              sx={{
                borderRadius: 2,
                justifyContent: sidebarCollapsed ? 'center' : 'flex-start',
                px: sidebarCollapsed ? 1.5 : 2,
                '&.Mui-selected': {
                  bgcolor: 'primary.main',
                  color: 'white',
                  '&:hover': {
                    bgcolor: 'primary.dark',
                  },
                  '& .MuiListItemIcon-root': {
                    color: 'white',
                  },
                },
                '&:hover': {
                  bgcolor: isDark ? 'grey.200' : 'grey.100',
                },
              }}
            >
              <ListItemIcon sx={{ minWidth: sidebarCollapsed ? 0 : 40, justifyContent: 'center' }}>
                {item.icon}
              </ListItemIcon>
              {!sidebarCollapsed && (
                <ListItemText 
                  primary={item.text} 
                  primaryTypographyProps={{ fontWeight: 500 }}
                />
              )}
            </ListItemButton>
          </ListItem>
        ))}
      </List>
      {!sidebarCollapsed && (
        <Box sx={{ p: 2 }}>
          <Typography variant="caption" color="text.disabled" display="block" textAlign="center">
            {t.app.copyright}
          </Typography>
        </Box>
      )}
    </Box>
  );

  return (
    <Box sx={{ display: 'flex', width: '100%' }}>
      <AppBar
        position="fixed"
        elevation={0}
        sx={{
          width: { md: `calc(100% - ${currentDrawerWidth}px)` },
          ml: { md: `${currentDrawerWidth}px` },
          bgcolor: 'background.paper',
          transition: theme.transitions.create(['width', 'margin'], {
            easing: theme.transitions.easing.sharp,
            duration: theme.transitions.duration.leavingScreen,
          }),
          borderBottom: '1px solid',
          borderColor: 'divider',
        }}
      >
        <Toolbar sx={{ justifyContent: 'space-between' }}>
          <Box sx={{ display: 'flex', alignItems: 'center' }}>
            <IconButton
              aria-label="open drawer"
              edge="start"
              onClick={handleDrawerToggle}
              sx={{ mr: 2, display: { md: 'none' }, color: 'text.primary' }}
            >
              <MenuIcon />
            </IconButton>
            <Typography variant="h6" noWrap component="div" color="text.primary" fontWeight={600}>
              {t.calendar.title}
            </Typography>
          </Box>
          <Box>
            <IconButton
              onClick={handleThemeMenuOpen}
              color="inherit"
              sx={{ 
                color: 'text.primary',
                bgcolor: isDark ? 'grey.200' : 'grey.100',
                '&:hover': {
                  bgcolor: isDark ? 'grey.300' : 'grey.200',
                },
              }}
            >
              {getThemeIcon()}
            </IconButton>
            <Menu
              anchorEl={themeMenuAnchor}
              open={Boolean(themeMenuAnchor)}
              onClose={handleThemeMenuClose}
              anchorOrigin={{
                vertical: 'bottom',
                horizontal: 'right',
              }}
              transformOrigin={{
                vertical: 'top',
                horizontal: 'right',
              }}
              PaperProps={{
                sx: {
                  mt: 1,
                  minWidth: 160,
                  borderRadius: 2,
                  border: '1px solid',
                  borderColor: 'divider',
                },
              }}
            >
              <MenuItem 
                onClick={() => handleThemeChange('light')}
                selected={mode === 'light'}
                sx={{ gap: 1.5 }}
              >
                <LightModeIcon fontSize="small" />
                {t.theme.light}
              </MenuItem>
              <MenuItem 
                onClick={() => handleThemeChange('dark')}
                selected={mode === 'dark'}
                sx={{ gap: 1.5 }}
              >
                <DarkModeIcon fontSize="small" />
                {t.theme.dark}
              </MenuItem>
              <MenuItem 
                onClick={() => handleThemeChange('system')}
                selected={mode === 'system'}
                sx={{ gap: 1.5 }}
              >
                <SystemModeIcon fontSize="small" />
                {t.theme.system}
              </MenuItem>
            </Menu>
          </Box>
        </Toolbar>
      </AppBar>
      <Box
        component="nav"
        sx={{ 
          width: { md: currentDrawerWidth }, 
          flexShrink: { md: 0 },
          transition: theme.transitions.create('width', {
            easing: theme.transitions.easing.sharp,
            duration: theme.transitions.duration.leavingScreen,
          }),
        }}
      >
        <Drawer
          variant="temporary"
          open={mobileOpen}
          onClose={handleDrawerToggle}
          ModalProps={{ keepMounted: true }}
          sx={{
            display: { xs: 'block', md: 'none' },
            '& .MuiDrawer-paper': { 
              boxSizing: 'border-box', 
              width: drawerWidth,
              border: 'none',
            },
          }}
        >
          {drawer}
        </Drawer>
        <Drawer
          variant="permanent"
          sx={{
            display: { xs: 'none', md: 'block' },
            '& .MuiDrawer-paper': { 
              boxSizing: 'border-box', 
              width: currentDrawerWidth,
              border: 'none',
              bgcolor: 'background.paper',
              borderRight: '1px solid',
              borderColor: 'divider',
              transition: theme.transitions.create('width', {
                easing: theme.transitions.easing.sharp,
                duration: theme.transitions.duration.leavingScreen,
              }),
              overflowX: 'hidden',
            },
          }}
          open
        >
          {drawer}
        </Drawer>
      </Box>
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          p: { xs: 2, md: 4 },
          width: { md: `calc(100% - ${currentDrawerWidth}px)` },
          mt: '64px',
          bgcolor: 'background.default',
          minHeight: 'calc(100vh - 64px)',
          transition: theme.transitions.create('width', {
            easing: theme.transitions.easing.sharp,
            duration: theme.transitions.duration.leavingScreen,
          }),
        }}
      >
        {children}
      </Box>
    </Box>
  );
};

export default Layout;
