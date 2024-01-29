import { useEffect, useMemo, useState, Dispatch, SetStateAction } from 'react';

import { useColorScheme } from '@mui/joy/styles';

import { AgGridReact } from 'ag-grid-react';
import { ColDef, ValueFormatterParams, ValueGetterParams } from 'ag-grid-community';
import "ag-grid-community/styles/ag-grid.css";
import "ag-grid-community/styles/ag-theme-quartz.css";

import { makeTable, tableFromIPC } from 'apache-arrow';

const initialColDefs = [
    {
        field: 'name',
        headerName: 'Name',
        filter: 'agTextColumnFilter',
        pinned: 'left',
    },
    {
        field: 'run_id',
        headerName: 'ID',
    },
    {
        field: 'experiment_name',
        headerName: 'Experiment',
        filter: 'agTextColumnFilter',
    },
    {
        field: 'creation_time',
        headerName: 'Date',
        // type: 'date',
        filter: 'agDateColumnFilter',
        valueFormatter: (params: ValueFormatterParams) => {
            return new Date(params.value).toLocaleString();
        },
    },
    {
        headerName: 'Duration',
        // type: 'timeColumn',
        filter: 'agNumberColumnFilter',
        valueGetter: (params: ValueGetterParams) => {
            return (new Date(params.data.end_time).getTime() - new Date(params.data.creation_time).getTime()) / 1000;
        },
        valueFormatter: (params: ValueFormatterParams) => {
            return `${Math.floor(params.value)}s ${Math.floor(params.value % 1 * 1000)}ms`;
        },
    },
    {
        field: 'archived',
        headerName: 'Archived',
        filter: 'agSetColumnFilter',
    },
    {
        field: 'active',
        headerName: 'Active',
        filter: 'agSetColumnFilter',
    },
] as ColDef[];

export default function Runs({
    selectedExperiments,
    searchQuery,
    setAvailableMetrics,
    setSelectedRuns,
}: {
    selectedExperiments: string[],
    searchQuery: string,
    setAvailableMetrics: Dispatch<SetStateAction<string[]>>,
    setSelectedRuns: Dispatch<SetStateAction<string[]>>,
}) {
    const { mode } = useColorScheme();
    const [rowData, setRowData] = useState(makeTable({}));
    const [colDefs, setColDefs] = useState(initialColDefs);
    const defaultColDef = useMemo(() => {
        return {
            flex: 1,
            minWidth: 150,
            menuTabs: ['filterMenuTab'],
        } as ColDef;
    }, []);

    useEffect(() => {
        const url = new URL('http://localhost:5000/aim/api/runs/search/run/arrow');
        const params = {
            // limit: '100',
            q: `run.experiment in [${selectedExperiments.map(e => `"${e}"`).join(',')}]`
        };
        if (searchQuery) {
            params.q += ` and (${searchQuery})`;
        }
        url.search = new URLSearchParams(params).toString();

        tableFromIPC(
            fetch(
                url,
            )
        ).then(table => {
            setColDefs([
                ...initialColDefs,
                ...table.schema.fields.filter(f => f.name.startsWith("metrics:")).map(f => {
                    return {
                        field: f.name,
                        headerName: f.name.replace(/^metrics:/, ""),
                        type: 'numericColumn',
                        filter: 'agNumberColumnFilter',
                        valueFormatter: (params: ValueFormatterParams) => {
                            return params.value.toFixed(2);
                        },
                    }
                }),
                ...table.schema.fields.filter(f => f.name.startsWith("params:")).map(f => {
                    return {
                        field: f.name,
                        headerName: f.name.replace(/^params:/, ""),
                        // type: 'textColumn',
                        filter: 'agTextColumnFilter',
                    }
                }),
                ...table.schema.fields.filter(f => f.name.startsWith("tags:")).map(f => {
                    return {
                        field: f.name,
                        headerName: f.name.replace(/^tags:/, ""),
                        // type: 'textColumn',
                        filter: 'agTextColumnFilter',
                    }
                }),
            ]);
            setRowData(table)
            setAvailableMetrics(table.schema.fields.filter(f => f.name.startsWith("metrics:")).map(f => f.name.replace(/^metrics:/, "")));
            setSelectedRuns(table.getChild("run_id")?.toArray() || []);
        });
    }, [selectedExperiments, searchQuery, setAvailableMetrics, setSelectedRuns]);

    return (
        <div
            className={`ag-theme-quartz${mode === 'dark' ? '-dark' : ''}`}
            style={{ width: '100%', height: '50%' }}
        >
            <AgGridReact
                rowData={rowData.toArray()}
                columnDefs={colDefs}
                defaultColDef={defaultColDef}
                autoSizeStrategy={{
                    type: 'fitCellContents',
                }}
            />
        </div>
    );
}
