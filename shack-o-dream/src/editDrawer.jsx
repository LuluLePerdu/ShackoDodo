import Dialog from '@mui/material/Dialog';
import {DialogTitle, IconButton, styled} from "@mui/material";
import * as React from 'react';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Drawer from '@mui/material/Drawer';
import { MdClose } from 'react-icons/md';
import {ReadyState} from "react-use-websocket";

export default function EditDrawer({ open, onClose, selectedRow, selectedItem, sendModifiedRequest, dropRequest, readyState }) {
    const [textValue, setTextValue] = React.useState('');

    React.useEffect(() => {
        if (selectedRow) {
            if (selectedRow && selectedRow.data) {
                setTextValue(JSON.stringify(selectedRow.data, null, 2));
            } else {
                setTextValue(JSON.stringify(selectedRow, null, 2));
            }
        }
    }, [selectedRow]);

    const handleSubmit = (e) => {
        e.preventDefault();

        console.log('handleSubmit called');
        console.log('selectedItem:', selectedItem);
        console.log('selectedRow:', selectedRow);
        console.log('sendModifiedRequest:', sendModifiedRequest);

        if (!selectedItem || !sendModifiedRequest) {
            console.log('Missing selectedItem or sendModifiedRequest');
            return;
        }

        try {
            const modifiedData = JSON.parse(textValue);
            console.log('Parsed modifiedData:', modifiedData);

            const requestData = modifiedData.data || modifiedData;

            const modifiedRequest = {
                id: selectedItem.id,
                method: requestData.method || 'GET',
                url: requestData.url || '',
                headers: requestData.headers || {},
                body: requestData.body || '',
                action: "send"
            };

            console.log('Sending modifiedRequest:', modifiedRequest);

            sendModifiedRequest(modifiedRequest);
            onClose();

        } catch (error) {
            console.error('Error in handleSubmit:', error);
            alert('Erreur dans le format JSON: ' + error.message);
        }
    };

    const DrawerHeader = styled('div')(({ theme }) => ({
        display: 'flex',
        alignItems: 'center',
        padding: theme.spacing(0, 1),
        // necessary for content to be below app bar
        ...theme.mixins.toolbar,
        justifyContent: 'flex-start',
    }));

    const isDisabled = readyState !== ReadyState.OPEN || selectedItem?.status !== 'pending';

    return (
        <>
            <Drawer
                sx={{
                    width: 600,
                    flexShrink: 0,
                    '& .MuiDrawer-paper': {
                        width: 600
                    },
                }}
                variant="persistent"
                anchor="right"
                open={open}
            >
                <DrawerHeader>
                    <Button color="cancel" onClick={onClose} sx={{ color: 'white' }}><MdClose/></Button>
                </DrawerHeader>

                {selectedItem?.status !== 'pending' && (
                    <div style={{
                        backgroundColor: '#fff3cd',
                        border: '1px solid #ffeaa7',
                        borderRadius: '4px',
                        padding: '15px',
                        margin: '15px 25px',
                        fontSize: '14px'
                    }}>
                        <strong>Attention:</strong> Cette requête ne peut pas être modifiée (statut: {selectedItem?.status}).
                        Seules les requêtes en statut "pending" peuvent être modifiées.
                    </div>
                )}

                <form onSubmit={handleSubmit} id="edit-request-form" style={{ marginRight: 25, marginLeft: 25}}>
                    <TextField
                        multiline
                        rows={20}
                        required
                        margin="dense"
                        id="req"
                        name="req"
                        label="Requête"
                        type="text"
                        fullWidth
                        disabled={isDisabled}
                        sx={{ '.MuiInputBase-input': {
                                //fontFamily: 'monospace',
                                fontFamily: '"Cascadia Code"'// Change font family
                            },
                        }}
                        value={textValue}
                        onChange={(e) => setTextValue(e.target.value)}
                        onKeyDown={(e) => {
                            if (e.key === 'Tab') {
                                e.preventDefault();

                                const textarea = e.target;
                                const start = textarea.selectionStart;
                                const end = textarea.selectionEnd;

                                const newValue =
                                    textValue.substring(0, start) +
                                    '\t' +
                                    textValue.substring(end);

                                setTextValue(newValue);

                                setTimeout(() => {
                                    textarea.selectionStart = textarea.selectionEnd = start + 1;
                                }, 0);
                            }
                        }}
                    />
                </form>
                <Button color="primary" type="submit" form="edit-request-form" disabled={isDisabled}>Modifier</Button>
            </Drawer>
        </>
    );
}
