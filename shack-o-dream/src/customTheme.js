import { createTheme } from '@mui/material/styles'
import { grey } from '@mui/material/colors';

const theme = createTheme({
    palette: {
        mode: 'dark',
        primary: {
            main: '#00FF00',
            dark: '#00AA00',
            light: '#66FF66',
            contrastText: '#000'
        },
        secondary: {
            main: '#8913CF',
            dark: '#5F0D90',
            light: '#A042D8',
            contrastText: '#fff'
        },
        cancel: {
            main: grey[700]
        },
        background: {
            default: '#1a1a1a',
            paper: '#2a2a2a',
        },
    }
});

export default theme;