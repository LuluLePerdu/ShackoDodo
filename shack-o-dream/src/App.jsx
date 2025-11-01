import './App.css'
import { FaPlay } from "react-icons/fa";
import { FaPause } from "react-icons/fa";

import * as React from 'react';
import Button from '@mui/material/Button';
import {Table, TableHead, TableRow} from "@mui/material";
import StickyHeadTable from "./tableaux.jsx";

function App() {

  return (
      <>
          <Button><FaPlay/></Button>
          <Button><FaPause/></Button>
          <Button>Nouvel onglet</Button>
          <StickyHeadTable />
          <script type="module" src="/src/tableaux.jsx"></script>
      </>
  )
}

export default App
