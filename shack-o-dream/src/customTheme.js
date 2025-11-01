import { createTheme } from '@mui/material/styles'


const theme = createTheme({
    palette: {
        primary: {
            main: '#00B600',
            dark: '#007F00',
            light: '#33C433',
            contrastText: '#000'
        },
        secondary: {
            main: '#8913CF',
            dark: '#5F0D90',
            light: '#A042D8',
            contrastText: '#fff'
        }
    }
});

export default theme;