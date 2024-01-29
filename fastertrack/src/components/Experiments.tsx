import { useCallback, useEffect, useState } from 'react';

import { useColorScheme } from '@mui/joy/styles';

import { AgGridReact } from 'ag-grid-react';
import { ColDef, FirstDataRenderedEvent, IRowNode, SelectionChangedEvent } from 'ag-grid-community';
import "ag-grid-community/styles/ag-grid.css";
import "ag-grid-community/styles/ag-theme-quartz.css";

export default function Experiments({ setSelectedExperiments }: { setSelectedExperiments: React.Dispatch<React.SetStateAction<string[]>> }) {
    const { mode } = useColorScheme();
    const [rowData, setRowData] = useState([]);
    const [colDefs] = useState([
        {
            field: 'name',
            headerName: "Experiments",
            headerCheckboxSelection: true,
            checkboxSelection: true,
            resizable: false,
            sort: 'asc',
        },
    ] as ColDef[]);

    useEffect(() => {
        fetch('http://localhost:5000/aim/api/experiments')
            .then(result => result.json())
            .then(rowData => setRowData(rowData));
    }, [])

    const onSelectionChanged = useCallback((event: SelectionChangedEvent) => {
        setSelectedExperiments(event.api.getSelectedNodes().map(node => node.data.name));
    }, [setSelectedExperiments]);

    const onFirstDataRendered = useCallback(
        (params: FirstDataRenderedEvent) => {
            const nodesToSelect: IRowNode[] = [];
            params.api.forEachNode((node: IRowNode) => {
                if (node.data && node.data.id === '0') {
                    nodesToSelect.push(node);
                }
            });
            params.api.setNodesSelected({ nodes: nodesToSelect, newValue: true });
        }, []);

    return (
        <div
            className={`ag-theme-quartz${mode === 'dark' ? '-dark' : ''}`}
            style={{ width: '100%', height: '100%' }}
        >
            <AgGridReact
                rowData={rowData}
                columnDefs={colDefs}
                rowSelection={'multiple'}
                onSelectionChanged={onSelectionChanged}
                onFirstDataRendered={onFirstDataRendered}
                autoSizeStrategy={{
                    type: 'fitGridWidth'
                }}
            />
        </div>
    );
}
