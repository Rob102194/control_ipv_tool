import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { Modal, Button, Spinner, Alert } from 'react-bootstrap';
import { useIPV } from '../../hooks/useIPV';
import ModeloIPV from './ModeloIPV';
import IPVRegistrosList from './IPVRegistrosList';
import ipvApi from '../../api/ipvApi';
import { formatDateAyer, formatDateEs } from '../../utils/date';
import { PlusIcon, ClockIcon, LayersIcon, iconoDeArea } from '../../components/icons';

function IPVControl() {
    const navigate = useNavigate();
    const { fecha, setFecha, inventario, loading, error, handleCargarDatos } = useIPV();

    const [showModeloModal, setShowModeloModal] = useState(false);
    const [showRegistrosListModal, setShowRegistrosListModal] = useState(false);
    const [registros, setRegistros] = useState([]);
    const [registrosLoading, setRegistrosLoading] = useState(true);

    // Por defecto, ayer: el IPV se revisa a día vencido.
    useEffect(() => {
        if (!fecha) setFecha(formatDateAyer());
    }, [fecha, setFecha]);

    // Carga el estado de todas las áreas para la fecha elegida (para las
    // tarjetas de "Estado de hoy"); es la misma llamada que antes solo se
    // hacía al abrir el registro.
    useEffect(() => {
        if (fecha) handleCargarDatos(fecha);
    }, [fecha, handleCargarDatos]);

    useEffect(() => {
        const cargarRegistros = async () => {
            setRegistrosLoading(true);
            try {
                const { data } = await ipvApi.getRegistros();
                const ordenados = [...data].sort((a, b) => b.fecha.localeCompare(a.fecha));
                setRegistros(ordenados);
            } catch (err) {
                console.error('Error al cargar los últimos registros', err);
            } finally {
                setRegistrosLoading(false);
            }
        };
        cargarRegistros();
    }, [showRegistrosListModal]);

    const areas = useMemo(() => Object.keys(inventario), [inventario]);
    const ultimosRegistros = registros.slice(0, 5);

    const irARegistro = useCallback((fechaDestino, areaNombre) => {
        const params = areaNombre ? `?area=${encodeURIComponent(areaNombre)}` : '';
        navigate(`/ipv/registro/${fechaDestino}${params}`);
    }, [navigate]);

    const handleSelectRegistro = (fechaSeleccionada) => {
        setShowRegistrosListModal(false);
        irARegistro(fechaSeleccionada);
    };

    return (
        <div className="container-xl py-4">
            <div className="mb-4">
                <h1 style={{ fontSize: '1.75rem', fontWeight: 700 }}>Control de Inventario Diario (IPV)</h1>
                <p className="mb-0" style={{ color: 'var(--text-secondary)' }}>
                    Selecciona una fecha para registrar o consultar el inventario del día.
                </p>
            </div>

            {error && <Alert variant="danger">{error}</Alert>}

            <div className="card d-flex flex-row flex-wrap align-items-end gap-3">
                <div>
                    <label className="form-label fw-bold text-uppercase" style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', letterSpacing: '0.03em' }} htmlFor="fechaSelector">
                        Fecha
                    </label>
                    <input
                        id="fechaSelector"
                        type="date"
                        className="form-control"
                        style={{ width: '180px' }}
                        value={fecha}
                        onChange={(e) => setFecha(e.target.value)}
                    />
                </div>
                <button
                    type="button"
                    className="btn btn-primary d-inline-flex align-items-center gap-2"
                    disabled={!fecha || loading}
                    onClick={() => irARegistro(fecha)}
                >
                    {loading ? <Spinner as="span" animation="border" size="sm" /> : <PlusIcon size={16} />}
                    Nuevo Registro Diario
                </button>
                <button
                    type="button"
                    className="btn btn-outline-secondary d-inline-flex align-items-center gap-2"
                    onClick={() => setShowRegistrosListModal(true)}
                >
                    <ClockIcon size={16} /> Consultar Registros
                </button>
                <button
                    type="button"
                    className="btn btn-link ms-auto d-inline-flex align-items-center gap-2 text-decoration-none"
                    style={{ color: 'var(--text-secondary)' }}
                    onClick={() => setShowModeloModal(true)}
                >
                    <LayersIcon size={16} /> Crear Modelos de Área
                </button>
            </div>

            <div className="mt-5">
                <h2 style={{ fontSize: '0.9375rem', fontWeight: 700 }}>
                    Estado de hoy{fecha ? ` · ${formatDateEs(fecha)}` : ''}
                </h2>
                {areas.length === 0 ? (
                    <p className="mt-2" style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
                        {loading ? 'Cargando áreas...' : 'No hay áreas con modelo de IPV configurado.'}
                    </p>
                ) : (
                    <div className="row row-cols-1 row-cols-sm-2 row-cols-lg-4 g-3 mt-1">
                        {areas.map(areaNombre => {
                            const Icon = iconoDeArea(areaNombre);
                            const cantidad = inventario[areaNombre].length;
                            return (
                                <div className="col" key={areaNombre}>
                                    <div className="card area-card h-100" onClick={() => irARegistro(fecha, areaNombre)}>
                                        <div className="area-icon-circle" style={{ backgroundColor: 'var(--color-primary-soft)', color: 'var(--color-primary)' }}>
                                            <Icon size={18} />
                                        </div>
                                        <div>
                                            <div className="fw-bold mb-1" style={{ fontSize: '0.9375rem' }}>{areaNombre}</div>
                                            <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>
                                                {cantidad} {cantidad === 1 ? 'producto' : 'productos'} en el modelo
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            );
                        })}
                    </div>
                )}
            </div>

            <div className="mt-5">
                <div className="d-flex align-items-center justify-content-between">
                    <h2 style={{ fontSize: '0.9375rem', fontWeight: 700 }}>Últimos registros</h2>
                    <button type="button" className="btn btn-link p-0 text-decoration-none" onClick={() => setShowRegistrosListModal(true)}>
                        Ver todos
                    </button>
                </div>
                <div className="card mt-2" style={{ padding: '4px 0' }}>
                    {registrosLoading ? (
                        <div className="text-center py-4">
                            <Spinner animation="border" size="sm" />
                        </div>
                    ) : ultimosRegistros.length === 0 ? (
                        <div className="text-center py-4" style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
                            Todavía no hay registros guardados.
                        </div>
                    ) : (
                        ultimosRegistros.map(registro => (
                            <div className="list-row" key={registro.fecha}>
                                <span style={{ fontWeight: 600, fontSize: '0.875rem' }}>{formatDateEs(registro.fecha)}</span>
                                <button
                                    type="button"
                                    className="btn btn-link p-0 fw-bold text-decoration-none"
                                    style={{ fontSize: '0.8125rem' }}
                                    onClick={() => handleSelectRegistro(registro.fecha)}
                                >
                                    Ver
                                </button>
                            </div>
                        ))
                    )}
                </div>
            </div>

            <Modal show={showModeloModal} onHide={() => setShowModeloModal(false)} size="lg">
                <Modal.Header closeButton>
                    <Modal.Title>Configuración de Modelos de IPV</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <ModeloIPV />
                </Modal.Body>
                <Modal.Footer>
                    <Button variant="secondary" onClick={() => setShowModeloModal(false)}>
                        Cerrar
                    </Button>
                </Modal.Footer>
            </Modal>

            <Modal show={showRegistrosListModal} onHide={() => setShowRegistrosListModal(false)} size="lg">
                <Modal.Header closeButton>
                    <Modal.Title>Consultar Registros de IPV</Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    <IPVRegistrosList onSelectRegistro={handleSelectRegistro} />
                </Modal.Body>
                <Modal.Footer>
                    <Button variant="secondary" onClick={() => setShowRegistrosListModal(false)}>
                        Cerrar
                    </Button>
                </Modal.Footer>
            </Modal>
        </div>
    );
}

export default IPVControl;
