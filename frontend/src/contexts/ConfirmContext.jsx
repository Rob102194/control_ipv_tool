import React, { createContext, useCallback, useContext, useState } from 'react';
import { Modal, Button } from 'react-bootstrap';
import { AlertTriangleIcon } from '../components/icons';

// Reemplaza window.confirm(): igual que el toast, el diálogo nativo se ve
// distinto por plataforma y no respeta el tema de la app. confirmar(...)
// devuelve una Promise<boolean>, usable como `if (await confirmar(msg))`.
const ConfirmContext = createContext(null);

export function ConfirmProvider({ children }) {
    const [pedido, setPedido] = useState(null); // { mensaje, resolve, confirmLabel, danger }

    const confirmar = useCallback((mensaje, opciones = {}) => {
        return new Promise((resolve) => {
            setPedido({ mensaje, resolve, ...opciones });
        });
    }, []);

    const resolver = (resultado) => {
        pedido?.resolve(resultado);
        setPedido(null);
    };

    return (
        <ConfirmContext.Provider value={confirmar}>
            {children}
            <Modal show={!!pedido} onHide={() => resolver(false)} centered>
                <Modal.Body className="d-flex gap-3 align-items-start pt-4">
                    <span
                        className="d-inline-flex align-items-center justify-content-center flex-shrink-0"
                        style={{
                            width: '40px', height: '40px', borderRadius: '50%',
                            backgroundColor: 'var(--warning-soft)', color: 'var(--warning-ink)',
                        }}
                    >
                        <AlertTriangleIcon size={20} />
                    </span>
                    <span style={{ paddingTop: '8px' }}>{pedido?.mensaje}</span>
                </Modal.Body>
                <Modal.Footer>
                    <Button variant="outline-secondary" onClick={() => resolver(false)}>
                        Cancelar
                    </Button>
                    <Button variant={pedido?.danger === false ? 'primary' : 'danger'} onClick={() => resolver(true)}>
                        {pedido?.confirmLabel || 'Eliminar'}
                    </Button>
                </Modal.Footer>
            </Modal>
        </ConfirmContext.Provider>
    );
}

// Devuelve confirmar(mensaje, { confirmLabel?, danger? }) -> Promise<boolean>.
export function useConfirm() {
    return useContext(ConfirmContext);
}
