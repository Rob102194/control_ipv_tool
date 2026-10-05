import 'bootstrap/dist/css/bootstrap.min.css';
import 'bootstrap/dist/js/bootstrap.bundle.min.js';
import ReactDOM from 'react-dom/client'
import './index.css';
import App from './App.jsx'
import { ThemeProvider } from './contexts/ThemeContext.jsx';
import { ToastProvider } from './contexts/ToastContext.jsx';
import { ConfirmProvider } from './contexts/ConfirmContext.jsx';
import { RestoreLockProvider } from './contexts/RestoreLockContext.jsx';

ReactDOM.createRoot(document.getElementById('root')).render(
  <ThemeProvider>
    <ToastProvider>
      <ConfirmProvider>
        <RestoreLockProvider>
          <App />
        </RestoreLockProvider>
      </ConfirmProvider>
    </ToastProvider>
  </ThemeProvider>
)
