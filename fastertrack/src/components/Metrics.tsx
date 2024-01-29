import { useEffect, useRef, useState } from "react";

import Stack from "@mui/joy/Stack";

import { makeTable, tableFromIPC } from 'apache-arrow';

import * as Plot from "@observablehq/plot";

import { fromArrow, op, escape } from 'arquero';

export default function Metrics({
    selectedMetrics,
    selectedRuns,
}: {
    selectedMetrics: string[],
    selectedRuns: string[],
}) {
    const [table, setTable] = useState(makeTable({}));
    const plotsRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (selectedRuns.length === 0 || selectedMetrics.length === 0) {
            return;
        }

        tableFromIPC(
            fetch(
                'http://localhost:5000/api/2.0/mlflow/metrics/get-histories',
                {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify({
                        'run_ids': selectedRuns,
                        'metric_keys': selectedMetrics,
                    }),
                }
            )
        ).then(table => {
            setTable(table);
        });
    }, [selectedMetrics, selectedRuns]);

    useEffect(() => {
        if (table.numRows === 0) {
            return;
        }

        const plots: ((HTMLElement | SVGSVGElement) & Plot.Plot)[] = [];

        const df = fromArrow(table);
        for (const metric of selectedMetrics) {
            const data = df.filter(escape((d: { key: string }) => op.equal(d.key, metric))).select('run_id', 'step', 'value');
            const plot = Plot.plot({
                title: metric,
                y: { grid: true },
                color: { scheme: 'spectral' },
                marks: [
                    Plot.lineY(data, { x: 'step', y: 'value', stroke: 'run_id', tip: true }),
                ],
            });
            plots.push(plot);
        }

        for (const plot of plots)
            plotsRef.current?.append(plot);

        return () => {
            for (const plot of plots)
                plot.remove();
        }
    });

    return (

        <Stack
            ref={plotsRef}
            direction={'row'}
            justifyContent={'center'}
            alignItems={'center'}
            spacing={2}
            flexWrap={'wrap'}
            sx={{
                height: '50%',
                width: '100%',
                overflow: 'auto',
            }}
        />

    )
}