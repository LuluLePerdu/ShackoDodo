import './App.css'
import { FaPlay } from "react-icons/fa";
import { FaPause } from "react-icons/fa";

import * as React from 'react';
import Button from '@mui/material/Button';
import StickyHeadTable from "./tableaux.jsx";
import theme from './customTheme.js';
import { ThemeProvider } from '@mui/material/styles';

function play() {}

function pause() {}

function newTab() {}


function App() {

  return (
      <ThemeProvider theme={theme}>
          <>
              <div className="top">
                  <div>
                    <Button onClick={play()}><FaPlay/></Button>
                    <Button onClick={pause()}><FaPause/></Button>
                  </div>
                  <Button onClick={newTab}>Nouvel onglet</Button>
              </div>
              <StickyHeadTable />
              <script type="module" src="/src/tableaux.jsx"></script>
          </>
      </ThemeProvider>
  )
}

export default App
