import React, { useState, useEffect } from 'react';
import { ListGroup, Button, Spinner, Alert } from 'react-bootstrap';
import ipvApi from '../../api/ipvApi';

function IPVRegistrosList({ onSelectRegistro }) {
    const [registros, setRegistros] = useState([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    useEffect(() => {
        const fetchRegistros = async () => {
            setLoading(true);
            setError('');
            try {
                const response = await ipvApi.getRegistros();
                setRegistros(response.data);
            } catch (err) {
                setError('Error al cargar los registros.');
                console.error(err);
            } finally {
                setLoading(false);
            }
        };
        fetchRegistros();
    }, []);

    return (
        <div>
            {loading && <Spinner animation="border" />}
            {error && <Alert variant="danger">{error}</Alert>}
            <ListGroup>
                {registros.length > 0 ? (
                    registros.map(registro => (
                        <ListGroup.Item key={registro.fecha}>
                            <div className="d-flex justify-content-between align-items-center">
                                <span>Registro del {registro.fecha}</span>
                                <Button variant="outline-primary" size="sm" onClick={() => onSelectRegistro(registro.fecha)}>
                                    Ver/Editar
                                </Button>
                            </div>
                        </ListGroup.Item>
                    ))
                ) : (
                    <ListGroup.Item>No se encontraron registros.</ListGroup.Item>
                )}
            </ListGroup>
        </div>
    );
}

export default IPVRegistrosList;
