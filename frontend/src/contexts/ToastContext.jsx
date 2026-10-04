import React, { createContext, useCallback, useContext, useRef, useState } from 'react';
import { CheckIcon, XCircleIcon, AlertTriangleIcon } from '../components/icons';

// Reemplaza alert(): los diálogos nativos del navegador/webview se ven
// distinto en cada plataforma (Windows/macOS/Linux, escritorio vs web), lo
// que rompe la identidad visual de una app multiplataforma. Este toast
// siempre se ve igual, con el tema claro/oscuro de la app.
const ToastContext = createContext(null);

const ICONS = { success: CheckIcon, danger: XCircleIcon, warning: AlertTriangleIcon };
const DURATION_MS = 4500;

export function ToastProvider({ children }) {
    const [toasts, setToasts] = useState([]);
    const timers = useRef(new Map());

    const dismiss = useCallback((id) => {
        setToasts(prev => prev.filter(t => t.id !== id));
        const timer = timers.current.get(id);
        if (timer) {
            clearTimeout(timer);
            timers.current.delete(id);
        }
    }, []);

    const showToast = useCallback((message, variant = 'success') => {
        const id = crypto.randomUUID();
        setToasts(prev => [...prev, { id, message, variant }]);
        const timer = setTimeout(() => dismiss(id), DURATION_MS);
        timers.current.set(id, timer);
    }, [dismiss]);

    return (
        <ToastContext.Provider value={showToast}>
            {children}
            <div className="app-toast-stack">
                {toasts.map(t => {
                    const Icon = ICONS[t.variant] || CheckIcon;
                    return (
                        <div key={t.id} className={`app-toast app-toast-${t.variant}`} role="status">
                            <Icon size={17} />
                            <span>{t.message}</span>
                            <button type="button" className="app-toast-close" onClick={() => dismiss(t.id)} aria-label="Cerrar notificación">
                                &times;
                            </button>
                        </div>
                    );
                })}
            </div>
        </ToastContext.Provider>
    );
}

// Devuelve showToast(mensaje, variante): variante es 'success' | 'danger' | 'warning'.
export function useToast() {
    return useContext(ToastContext);
}
