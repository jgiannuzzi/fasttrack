import { useState } from 'react';
import { CssVarsProvider } from '@mui/joy/styles';
import CssBaseline from '@mui/joy/CssBaseline';

import './App.css'
import ColorSchemeDetector from './components/ColorSchemeDetector';
import Layout from './components/Layout';
import Experiments from './components/Experiments';
import Header from './components/Header';
import Main from './components/Main';

export default function App() {
  const [selectedExperiments, setSelectedExperiments] = useState([] as string[]);

  return (
    <CssVarsProvider disableTransitionOnChange>
      <ColorSchemeDetector />
      <CssBaseline />
      <Layout.Root>
        <Layout.Header>
          <Header />
        </Layout.Header>
        <Layout.SideNav>
          <Experiments setSelectedExperiments={setSelectedExperiments} />
        </Layout.SideNav>
        <Layout.Main>
          <Main selectedExperiments={selectedExperiments} />
        </Layout.Main>
      </Layout.Root>
    </CssVarsProvider>
  );
}