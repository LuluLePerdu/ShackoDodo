import './App.css'
import {FaForward, FaPlay} from "react-icons/fa";
import { FaPause } from "react-icons/fa";

import * as React from 'react';
import Button from '@mui/material/Button';
import StickyHeadTable from "./tableaux.jsx";
import theme from './customTheme.js';
import { ThemeProvider } from '@mui/material/styles';
import useWebSocket, {ReadyState} from "react-use-websocket";
import {Box, Typography} from "@mui/material";


function App() {
    const WS_URL = `ws://127.0.0.1:8182/ws`;
    const { sendMessage, lastMessage, readyState } = useWebSocket(WS_URL);
    const [items, setItems] = React.useState([]);

    React.useEffect(() => {
        if (lastMessage !== null) {
            try {
                const parsed = JSON.parse(lastMessage.data);
                addItem(createData(Date.now(), parsed.data.url, parsed.data.method, "", "", "", parsed));
            } catch(err) {
                console.error("Error parsing WebSocket message:", err, lastMessage.data);
            }
        }
    }, [lastMessage]);

    function clear(){
        setItems([]);
    }
    function play() {}

    function pause() {}

    function foward() {}

    function newTab() {}

    const connectionStatus = {
        [ReadyState.CONNECTING]: 'En connexion...',
        [ReadyState.OPEN]: 'Ouverte',
        [ReadyState.CLOSING]: 'Fermeture...',
        [ReadyState.CLOSED]: 'Fermé',
        [ReadyState.UNINSTANTIATED]: 'Non instancié',
    }[readyState];

    const handleClickSendMessage = () => {
        if (readyState === ReadyState.OPEN) {
            sendMessage('Hello from React!');
        }
    };

    function createData(id, url, method, path, query, status, data) {
        return { id, url, method, path, query, status, data };
    }

    const addItem = (newItem) => {
        setItems([newItem, ...items]);
    };

    const removeItem = (idToRemove) => {
        setItems(items.filter(item => item.id !== idToRemove));
    };

    const getDotColor = () => {
        switch (readyState) {
            case ReadyState.OPEN:
                return "green";
            case ReadyState.CONNECTING:
                return "yellow";
            case ReadyState.CLOSED:
            case ReadyState.CLOSING:
                return "red";
            default:
                return "gray";
        }
    };


  return (
      <ThemeProvider theme={theme}>
          <>
              <div className="top">
                  <div>
                      <Box p={2}>
                          <Box display="flex" alignItems="center" gap={1}>
                              <Box
                                  sx={{
                                      width: 10,
                                      height: 10,
                                      borderRadius: "50%",
                                      backgroundColor: getDotColor(),
                                  }}
                              />
                              <Typography variant="body2">
                                  Connexion: {connectionStatus}
                              </Typography>
                          </Box>
                      </Box>
                      <Button onClick={play}><FaPlay/></Button>
                      <Button onClick={pause}><FaPause/></Button>
                      <Button onClick={foward}><FaForward/></Button>
                  </div>
                  <Button onClick={clear}>Supprimer toutes les requêtes</Button>
                  <Button onClick={newTab}>Nouvel onglet</Button>
              </div>
              <StickyHeadTable items={items} handleDeleteItem={removeItem}/>
              <script type="module" src="/src/tableaux.jsx"></script>
          </>
      </ThemeProvider>
  )
}

export default App
