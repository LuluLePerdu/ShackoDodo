import * as React from 'react';
import Paper from '@mui/material/Paper';
import Table from '@mui/material/Table';
import TableBody from '@mui/material/TableBody';
import TableCell from '@mui/material/TableCell';
import TableContainer from '@mui/material/TableContainer';
import TableHead from '@mui/material/TableHead';
import TablePagination from '@mui/material/TablePagination';
import TableRow from '@mui/material/TableRow';
import EditDrawer from "./editDrawer.jsx";
import DeleteIcon from '@mui/icons-material/Delete';
import IconButton from '@mui/material/IconButton';


const columns = [
    { id: 'url', label: 'URL', minWidth: 100 },
    { id: 'method', label: 'Méthode', minWidth: 75},
    { id: 'status', label: 'Statut', minWidth: 75}
];

function createData(id, url, method, path, query, status) {
    return { id, url, method, path, query, status };
}

export default function StickyHeadTable({items, handleDeleteItem, sendModifiedRequest, readyState}) {
    const [page, setPage] = React.useState(0);
    const [rowsPerPage, setRowsPerPage] = React.useState(10);
    const [selectedRow, setSelectedRow] = React.useState(null);
    const [drawerOpen, setDrawerOpen] = React.useState(false);


    const handleChangePage = (event, newPage) => {
        setPage(newPage);
    };

    const handleChangeRowsPerPage = (event) => {
        setRowsPerPage(+event.target.value);
        setPage(0);
    };

    const handleRowClick = (row) => {
        console.log('Row clicked:', row);
        setSelectedRow(row);
        setDrawerOpen(true);
    };

    const handleDrawerClose = () => {
        setDrawerOpen(false);
        setSelectedRow(null)
    };

    return (
        <Paper sx={{ width: '100%', overflow: 'hidden' }}>
            <TableContainer sx={{ maxHeight: '70vh',
                overflowX: 'auto',
<<<<<<< HEAD
                transformOrigin: 'left center',
=======
>>>>>>> 994ae25a28171b3edb442a5b630adade5305c2cc
            }}>
                <Table stickyHeader aria-label="sticky table" size="small">
                    <TableHead>
                        <TableRow>
                            {columns.map((column) => (
                                <TableCell
                                    key={column.id}
                                    align={column.align}
                                    style={{ minWidth: column.minWidth }}
                                >
                                    {column.label}
                                </TableCell>
                            ))}
                            <TableCell></TableCell>
                        </TableRow>
                    </TableHead>
                    <TableBody>
                        {items
                            .slice(page * rowsPerPage, page * rowsPerPage + rowsPerPage)
                            .map((row) => {
                                return (
                                    <TableRow hover role="checkbox" tabIndex={-1} key={row.id} >
                                        {columns.map((column) => {
                                            const value = row[column.id];
                                            return (
                                                <TableCell key={column.id} align={column.align} onClick={() => handleRowClick(row)}>
                                                    {column.id === 'url' && typeof value === 'string' && value.length > 50
                                                        ? `${value.substring(0, 100)}...`
                                                        : column.format && typeof value === 'number'
                                                            ? column.format(value)
                                                            : value}
                                                </TableCell>
                                            );
                                        })}
                                        <TableCell><IconButton onClick={() => handleDeleteItem(row.id)}><DeleteIcon/></IconButton></TableCell>
                                    </TableRow>
                                );
                            })}
                    </TableBody>
                </Table>
            </TableContainer>

            <TablePagination
                rowsPerPageOptions={[10, 25, 100]}
                component="div"
                count={items.length}
                rowsPerPage={rowsPerPage}
                page={page}
                onPageChange={handleChangePage}
                onRowsPerPageChange={handleChangeRowsPerPage}
            />
            <EditDrawer
                open={drawerOpen}
                onClose={handleDrawerClose}
                selectedRow={selectedRow != null ? selectedRow.data : ''}
                selectedItem={selectedRow}
                sendModifiedRequest={sendModifiedRequest}
                readyState={readyState}
            />
        </Paper>

    );
}