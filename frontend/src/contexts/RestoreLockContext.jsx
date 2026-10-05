import React, { createContext, useContext, useState } from 'react';
import { CheckIcon } from '../components/icons';

// Tras restaurar un backup, el backend cierra la conexión a la BD a
// propósito (ver internal/appboot/backup.go: no se intenta un hot-swap de la
// conexión compartida con todos los casos de uso, es más seguro pedir un
// reinicio limpio). Desde ese momento CUALQUIER llamada a la API fallaría de
// forma confusa, así que se bloquea toda la app con una pantalla fija —
// vive a nivel de la app entera (no de la página de Configuración) para que
// cubra la pantalla aunque el usuario navegue a otra ruta.
const RestoreLockContext = createContext(null);

// sessionStorage (no solo estado de React): si el usuario recarga la página
// a mano (F5) tras ver el aviso, la pantalla de bloqueo debe seguir ahí —
// sin esto, la recarga perdería el estado en memoria y el usuario vería los
// errores de conexión genéricos de cada vista en vez de un mensaje claro.
// sessionStorage se limpia solo al cerrar la ventana/pestaña, justo cuando
// corresponde que el bloqueo desaparezca (reabrir = reiniciar la app).
const STORAGE_KEY = 'controlipv-restored';

function leerBloqueadoInicial() {
    try {
        return sessionStorage.getItem(STORAGE_KEY) === '1';
    } catch {
        return false; // sessionStorage puede no estar disponible (modo privado, etc.)
    }
}

export function RestoreLockProvider({ children }) {
    const [bloqueado, setBloqueado] = useState(leerBloqueadoInicial);

    const bloquear = () => {
        try {
            sessionStorage.setItem(STORAGE_KEY, '1');
        } catch {
            // si falla, el bloqueo sigue funcionando en memoria para esta sesión
        }
        setBloqueado(true);
    };

    return (
        <RestoreLockContext.Provider value={bloquear}>
            {children}
            {bloqueado && (
                <div
                    style={{
                        position: 'fixed', inset: 0, zIndex: 3000,
                        backgroundColor: 'var(--color-base)',
                        display: 'flex', alignItems: 'center', justifyContent: 'center',
                        padding: '24px',
                    }}
                >
                    <div className="card" style={{ maxWidth: '420px', textAlign: 'center' }}>
                        <div
                            className="d-inline-flex align-items-center justify-content-center mx-auto mb-3"
                            style={{
                                width: '48px', height: '48px', borderRadius: '50%',
                                backgroundColor: 'var(--color-tertiary-soft)', color: 'var(--color-tertiary)',
                            }}
                        >
                            <CheckIcon size={22} />
                        </div>
                        <h2 style={{ fontSize: '1.125rem', fontWeight: 700, marginBottom: '8px' }}>
                            Base de datos restaurada
                        </h2>
                        <p style={{ color: 'var(--text-secondary)', marginBottom: 0 }}>
                            Cierra esta ventana y vuelve a abrir la aplicación para continuar.
                        </p>
                    </div>
                </div>
            )}
        </RestoreLockContext.Provider>
    );
}

// Devuelve una función sin argumentos: llamarla bloquea la app entera.
export function useRestoreLock() {
    return useContext(RestoreLockContext);
}
