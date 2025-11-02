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
    const [isPaused, setIsPaused] = React.useState(false);

    React.useEffect(() => {
        if (lastMessage !== null) {
            try {
                const parsed = JSON.parse(lastMessage.data);
                addItem(createData(parsed.id, parsed.data.url, parsed.data.method, "", "", parsed.data.status || "passthrough", parsed));
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

    function newTab() {
        if (readyState === ReadyState.OPEN) {
            sendMessage(JSON.stringify({
                type: 'launch_browser',
                data: {
                    browser: 'firefox'
                }
            }));
        }
    }

    function sendModifiedRequest(modifiedRequest) {
        console.log('sendModifiedRequest called with:', modifiedRequest);
        console.log('readyState:', readyState);

        if (readyState === ReadyState.OPEN) {
            // Extract ID from modifiedRequest and send it at message level
            const { id, ...requestData } = modifiedRequest;
            const message = JSON.stringify({
                type: 'modify_request',
                id: id,
                data: requestData
            });

            console.log('Sending WebSocket message:', message);
            sendMessage(message);
            updateItemStatus(id, 'sent');
            console.log('Item status updated to sent for id:', id);
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
                      <Button
                          onClick={play}
                          variant="text"
                          disabled={readyState !== ReadyState.OPEN}
                          sx={{
                              backgroundColor: 'transparent',
                              color: !isPaused ? '#000000' : '#4caf50',
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
                      <Button
                          onClick={pause}
                          variant="text"
                          disabled={readyState !== ReadyState.OPEN}
                          sx={{
                              backgroundColor: 'transparent',
                              color: isPaused ? '#000000' : '#4caf50',
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
                      <Button
                          onClick={foward}
                          variant="text"
                          disabled={readyState !== ReadyState.OPEN}
                          title="Envoyer toutes les requêtes en attente"
                          sx={{
                              backgroundColor: 'transparent',
                              color: '#4caf50',
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
                  </div>
                  <Button onClick={clear}>Supprimer toutes les requêtes</Button>
                  <Button
                      onClick={newTab}
                      disabled={readyState !== ReadyState.OPEN}
                  >
                      Nouvel onglet
                  </Button>
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
