import * as React from 'react';
import PropTypes from 'prop-types';
import Button from '@mui/material/Button';
import List from '@mui/material/List';
import ListItem from '@mui/material/ListItem';
import ListItemButton from '@mui/material/ListItemButton';
import ListItemText from '@mui/material/ListItemText';
import DialogTitle from '@mui/material/DialogTitle';
import Dialog from '@mui/material/Dialog';

export function BrowserDialog({ open, onClose }) {
    const browsers = ["firefox", "chrome", "edge"];

    const handleSelect = (browser) => {
        onClose(browser);
    };

    return (
        <Dialog onClose={() => onClose(null)} open={open}>
            <DialogTitle>Choisir un navigateur</DialogTitle>
            <List sx={{ pt: 0 }}>
                {browsers.map((browser) => (
                    <ListItem disablePadding key={browser}>
                        <ListItemButton onClick={() => handleSelect(browser)}>
                            <ListItemText primary={browser.charAt(0).toUpperCase() + browser.slice(1)} />
                        </ListItemButton>
                    </ListItem>
                ))}
            </List>
        </Dialog>
    );
}

BrowserDialog.propTypes = {
    open: PropTypes.bool.isRequired,
    onClose: PropTypes.func.isRequired,
};