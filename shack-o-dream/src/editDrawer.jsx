import Dialog from '@mui/material/Dialog';
import {DialogTitle, IconButton, styled} from "@mui/material";
import * as React from 'react';
import Button from '@mui/material/Button';
import TextField from '@mui/material/TextField';
import Drawer from '@mui/material/Drawer';
import { MdClose } from 'react-icons/md';

export default function EditDrawer({ open, onClose, selectedRow }) {
    const handleSubmit = (e) => {
        e.preventDefault();
    }

    const [textValue, setTextValue] = React.useState('');
    React.useEffect(() => {
        if (selectedRow) {
            setTextValue(JSON.stringify(selectedRow, null, 2));
        }
    }, [selectedRow]);

    const DrawerHeader = styled('div')(({ theme }) => ({
        display: 'flex',
        alignItems: 'center',
        padding: theme.spacing(0, 1),
        // necessary for content to be below app bar
        ...theme.mixins.toolbar,
        justifyContent: 'flex-start',
    }));

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
                <Button color="primary" type="submit" form="edit-request-form">Modifier</Button>
            </Drawer>
        </>
    );
}
