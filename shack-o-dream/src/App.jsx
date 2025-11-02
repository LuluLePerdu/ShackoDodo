import './App.css'
import { FaPlay } from "react-icons/fa";
import { FaPause } from "react-icons/fa";

import * as React from 'react';
import Button from '@mui/material/Button';
import StickyHeadTable from "./tableaux.jsx";
import theme from './customTheme.js';
import { ThemeProvider } from '@mui/material/styles';
import useWebSocket, {ReadyState} from "react-use-websocket";


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

    function newTab() {}

    const connectionStatus = {
        [ReadyState.CONNECTING]: 'Connecting',
        [ReadyState.OPEN]: 'Open',
        [ReadyState.CLOSING]: 'Closing',
        [ReadyState.CLOSED]: 'Closed',
        [ReadyState.UNINSTANTIATED]: 'Uninstantiated',
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


  return (
      <ThemeProvider theme={theme}>
          <>
              <div className="top">
                  <div>
                    <Button onClick={play}><FaPlay/></Button>
                    <Button onClick={pause}><FaPause/></Button>
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
