import './App.css'
import {FaForward, FaPlay} from "react-icons/fa";
import { FaPause } from "react-icons/fa";

import * as React from 'react';
import Button from '@mui/material/Button';
import StickyHeadTable from "./tableaux.jsx";
import theme from './customTheme.js';
import { ThemeProvider } from '@mui/material/styles';
import useWebSocket, {ReadyState} from "react-use-websocket";
import {Box, Tooltip, Typography} from "@mui/material";
import {BrowserDialog} from "./dialog.jsx";
import {grey} from "@mui/material/colors";


function App() {
    const WS_URL = `ws://127.0.0.1:8182/ws`;
    const { sendMessage, lastMessage, readyState } = useWebSocket(WS_URL);
    const [items, setItems] = React.useState([]);
    const [isPaused, setIsPaused] = React.useState(false);
    const [browserDialogOpen, setBrowserDialogOpen] = React.useState(false);

    React.useEffect(() => {
        if (lastMessage !== null) {
            try {
                const parsed = JSON.parse(lastMessage.data);
                addItem(createData(parsed.data.id, parsed.data.url, parsed.data.method, "", "", parsed.data.status || "passthrough", parsed));
            } catch(err) {
                console.error("Error parsing WebSocket message:", err, lastMessage.data);
            }
        }
    }, [lastMessage]);

    function clear(){
        setItems([]);
    }

    function play() {
        if (readyState === ReadyState.OPEN) {
            setIsPaused(false);
            sendMessage(JSON.stringify({
                type: 'pause',
                data: false
            }));

            setItems(prevItems =>
                prevItems.map(item => ({
                    ...item,
                    status: item.status === 'pending' ? 'sent' : item.status
                }))
            );
        }
    }

    function pause() {
        if (readyState === ReadyState.OPEN) {
            setIsPaused(true);
            sendMessage(JSON.stringify({
                type: 'pause',
                data: true
            }));
        }
    }

    function foward() {
        if (readyState === ReadyState.OPEN) {
            sendMessage(JSON.stringify({
                type: 'resume_all',
                data: true
            }));

            setItems(prevItems =>
                prevItems.map(item => ({
                    ...item,
                    status: item.status === 'pending' ? 'sent' : item.status
                }))
            );
        }
    }

    function sendModifiedRequest(modifiedRequest) {
        console.log('sendModifiedRequest called with:', modifiedRequest);
        console.log('readyState:', readyState);

        if (readyState === ReadyState.OPEN) {
            // Utiliser exactement le même format que payload-modifier.html
            const message = JSON.stringify({
                type: 'modify_request',
                data: modifiedRequest  // Envoyer directement l'objet modifiedRequest
            });

            console.log('Sending WebSocket message:', message);
            sendMessage(message);
            updateItemStatus(modifiedRequest.id, 'sent');
            console.log('Item status updated to sent for id:', modifiedRequest.id);
        } else {
            console.log('WebSocket not open, readyState:', readyState);
        }
    }

    function updateItemStatus(id, newStatus) {
        setItems(prevItems =>
            prevItems.map(item =>
                item.id === id ? {...item, status: newStatus} : item
            )
        );
    }
    
    function launchBrowser(browserName) {
        if (readyState === ReadyState.OPEN) {
            sendMessage(JSON.stringify({
                type: 'launch_browser',
                data: {
                    browser: browserName
                }
            }));
        }
    }

    const handleOpenBrowserDialog = () => {
        setBrowserDialogOpen(true);
    };


    const handleBrowserDialogClose = (browser) => {
        setBrowserDialogOpen(false);
        if (browser) {
            launchBrowser(browser);
        }
    };


    function newTab() {
        handleOpenBrowserDialog();
    }

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
                      <Tooltip title="Reprendre la réception de requêtes">
                          <Button
                              onClick={play}
                              variant="text"
                              disabled={readyState !== ReadyState.OPEN}
                              sx={{
                                  backgroundColor: 'transparent',
                                  color: !isPaused ? grey[500] : '#00FF00',
                                  border: 'none',
                                  boxShadow: 'none',
                                  '&:hover': {
                                      backgroundColor: 'rgba(0,0,0,0.04)',
                                      boxShadow: 'none'
                                  }
                              }}
                          >
                              <FaPlay/>
                          </Button>
                      </Tooltip>
                      <Tooltip title="Mettre sur pause la réceptions de requêtes">
                          <Button
                              onClick={pause}
                              variant="text"
                              disabled={readyState !== ReadyState.OPEN}
                              sx={{
                                  backgroundColor: 'transparent',
                                  color: isPaused ? grey[500] : '#00FF00',
                                  border: 'none',
                                  boxShadow: 'none',
                                  '&:hover': {
                                      backgroundColor: 'rgba(0,0,0,0.04)',
                                      boxShadow: 'none'
                                  }
                              }}
                          >
                              <FaPause/>
                          </Button>
                      </Tooltip>
                      <Tooltip title="Envoyer toutes les requêtes en attente">
                          <Button
                              onClick={foward}
                              variant="text"
                              disabled={readyState !== ReadyState.OPEN}
                              title="Envoyer toutes les requêtes en attente"
                              sx={{
                                  backgroundColor: 'transparent',
                                  color: '#00FF00',
                                  border: 'none',
                                  boxShadow: 'none',
                                  '&:hover': {
                                      backgroundColor: 'rgba(0,0,0,0.04)',
                                      boxShadow: 'none'
                                  }
                              }}
                          >
                              <FaForward/>
                          </Button>
                      </Tooltip>
                  </div>
                  <Tooltip title="Supprimer toutes les requêtes">
                    <Button disabled={readyState !== ReadyState.OPEN || items.length === 0} onClick={clear}>Supprimer toutes les requêtes</Button>
                  </Tooltip>
                  <Tooltip title="Ouvrir une nouvelle connexion">
                    <Button disabled={readyState !== ReadyState.OPEN} onClick={newTab}>Nouvelle connexion</Button>
                  </Tooltip>
                  <BrowserDialog
                      open={browserDialogOpen}
                      onClose={handleBrowserDialogClose}
                  />
              </div>
              <StickyHeadTable
                  items={items}
                  handleDeleteItem={removeItem}
                  sendModifiedRequest={sendModifiedRequest}
                  readyState={readyState}
              />
              <script type="module" src="/src/tableaux.jsx"></script>
          </>
      </ThemeProvider>
  )
}

export default App
