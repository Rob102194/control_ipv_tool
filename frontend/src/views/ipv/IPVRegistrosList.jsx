import React, { useState, useEffect } from 'react';
import { Button, Spinner, Alert } from 'react-bootstrap';
import ipvApi from '../../api/ipvApi';
import { formatDateEs } from '../../utils/date';

function IPVRegistrosList({ onSelectRegistro }) {
    const [registros, setRegistros] = useState([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    const fetchRegistros = async () => {
        setLoading(true);
        setError('');
        try {
            const response = await ipvApi.getRegistros();
            const ordenados = [...response.data].sort((a, b) => b.fecha.localeCompare(a.fecha));
            setRegistros(ordenados);
        } catch (err) {
            setError('Error al cargar los registros.');
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchRegistros();
    }, []);

    if (error) {
        return (
            <Alert variant="danger">
                {error}{' '}
                <Button variant="link" className="p-0 align-baseline" onClick={fetchRegistros}>Reintentar</Button>
            </Alert>
        );
    }

    return (
        <div className="card" style={{ padding: '4px 0' }}>
            {loading ? (
                <div className="text-center py-4">
                    <Spinner animation="border" size="sm" />
                </div>
            ) : registros.length === 0 ? (
                <div className="text-center py-4" style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
                    Todavía no hay registros guardados.
                </div>
            ) : (
                registros.map(registro => (
                    <div className="list-row" key={registro.fecha}>
                        <span style={{ fontWeight: 600, fontSize: '0.875rem' }}>{formatDateEs(registro.fecha)}</span>
                        <button
                            type="button"
                            className="btn btn-link p-0 fw-bold text-decoration-none"
                            style={{ fontSize: '0.8125rem' }}
                            onClick={() => onSelectRegistro(registro.fecha)}
                        >
                            Ver/Editar
                        </button>
                    </div>
                ))
            )}
        </div>
    );
}

export default IPVRegistrosList;
