import React, { useState, useEffect } from 'react';
import { Button, Container, Alert, Spinner, Form, Row, Col } from 'react-bootstrap';
import { getVentas, updateVenta, deleteVenta, importVentas, deleteVentas } from '../../api/ventaApi';
import { formatDateLocal } from '../../utils/date';
import { useToast } from '../../contexts/ToastContext';
import { useConfirm } from '../../contexts/ConfirmContext';
import { PencilIcon, TrashIcon, SearchIcon, CheckIcon } from '../../components/icons';

// Componente principal para la gestión de ventas.
const VentaList = () => {
    // Estado para gestionar la vista actual ('importar' o 'consultar').
    const [view, setView] = useState('importar');

    // Renderiza el componente principal y redirige a la vista seleccionada.
    return (
        <Container className="mt-4">
            <div className="mb-4">
                <h1 style={{ fontSize: '1.75rem', fontWeight: 700, marginBottom: '6px' }}>Ventas</h1>
                <p style={{ margin: 0, color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
                    Importa las ventas del día o consulta y corrige lo ya importado.
                </p>
            </div>
            <div style={{ display: 'inline-flex', gap: '4px', backgroundColor: 'var(--color-base)', border: '1px solid var(--border-color)', borderRadius: '10px', padding: '4px', marginBottom: '24px' }}>
                <button
                    type="button"
                    className={`ipv-tab${view === 'importar' ? ' active' : ''}`}
                    onClick={() => setView('importar')}
                >
                    Importar Ventas
                </button>
                <button
                    type="button"
                    className={`ipv-tab${view === 'consultar' ? ' active' : ''}`}
                    onClick={() => setView('consultar')}
                >
                    Consultar Ventas
                </button>
            </div>
            {view === 'importar' ? <ImportarVentas /> : <ConsultarVentas />}
        </Container>
    );
};

// Componente para la importación de ventas.
const ImportarVentas = () => {
    const showToast = useToast();
    const [file, setFile] = useState(null);
    const [fecha, setFecha] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [resultado, setResultado] = useState(null);

    // Maneja la importación del archivo de ventas.
    const handleImport = async () => {
        if (!file || !fecha) {
            showToast('Por favor, seleccione un archivo y una fecha.', 'warning');
            return;
        }
        setLoading(true);
        setError('');
        setResultado(null);
        try {
            const response = await importVentas(file, fecha);
            setResultado(response.data);
            setFile(null);
            setFecha('');
        } catch (err) {
            setError(err.response?.data?.error || 'Error al importar las ventas');
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="card" style={{ maxWidth: '720px' }}>
            <h2 style={{ fontSize: '1.0625rem', fontWeight: 700, marginBottom: '16px' }}>Importar Ventas desde Excel</h2>
            {error && <Alert variant="danger">{error}</Alert>}
            {resultado && (
                <Alert variant="success" onClose={() => setResultado(null)} dismissible>
                    <div>{resultado.message}</div>
                    {resultado.nuevas_recetas?.length > 0 && (
                        <div className="mt-2">
                            Se crearon {resultado.nuevas_recetas.length} receta{resultado.nuevas_recetas.length === 1 ? '' : 's'} nueva{resultado.nuevas_recetas.length === 1 ? '' : 's'} automáticamente
                            (sin ingredientes todavía): {resultado.nuevas_recetas.map((r) => r.nombre).join(', ')}.
                        </div>
                    )}
                </Alert>
            )}
            <Row>
                <Col md={6}>
                    <Form.Group>
                        <Form.Label className="fw-semibold">Archivo Excel</Form.Label>
                        <Form.Control type="file" accept=".xlsx, .xls" onChange={(e) => setFile(e.target.files[0])} />
                    </Form.Group>
                </Col>
                <Col md={6}>
                    <Form.Group>
                        <Form.Label className="fw-semibold">Fecha de las Ventas</Form.Label>
                        <Form.Control type="date" value={fecha} onChange={(e) => setFecha(e.target.value)} />
                    </Form.Group>
                </Col>
            </Row>
            <Button variant="primary" onClick={handleImport} className="mt-3" disabled={!file || !fecha || loading}>
                {loading ? <><Spinner as="span" animation="border" size="sm" /> Importando...</> : 'Importar'}
            </Button>
        </div>
    );
};

// Componente para consultar, editar y eliminar ventas.
const ConsultarVentas = () => {
    const confirmar = useConfirm();
    const [ventasOriginales, setVentasOriginales] = useState([]);
    const [ventas, setVentas] = useState([]);
    const [filtroNombre, setFiltroNombre] = useState('');
    const [fechaConsulta, setFechaConsulta] = useState(formatDateLocal(new Date()));
    const [editingId, setEditingId] = useState(null);
    const [editedData, setEditedData] = useState({});
    const [selectedIds, setSelectedIds] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    // Carga las ventas al montar el componente (y permite reintentar tras un error).
    const [reloadToken, setReloadToken] = useState(0);
    useEffect(() => {
        const fetchVentas = async () => {
            setLoading(true);
            setError('');
            try {
                const response = await getVentas();
                setVentasOriginales(response.data);
            } catch (err) {
                console.error(err);
                setError('Error al cargar las ventas');
            } finally {
                setLoading(false);
            }
        };
        fetchVentas();
    }, [reloadToken]);

    // Filtra las ventas por fecha y nombre de receta.
    useEffect(() => {
        let ventasFiltradas = ventasOriginales.filter(v => v.fecha === fechaConsulta);
        if (filtroNombre) {
            ventasFiltradas = ventasFiltradas.filter(v =>
                v.receta_nombre.toLowerCase().includes(filtroNombre.toLowerCase())
            );
        }
        setVentas(ventasFiltradas);
    }, [fechaConsulta, filtroNombre, ventasOriginales]);

    // Inicia el modo de edición.
    const handleEdit = (venta) => {
        setEditingId(venta.id);
        setEditedData({ ...venta });
    };

    // Guarda los cambios de una venta.
    const handleSave = async (id) => {
        try {
            await updateVenta(id, editedData);
            setEditingId(null);
            const updatedVentas = ventasOriginales.map(v => v.id === id ? editedData : v);
            setVentasOriginales(updatedVentas);
            setVentas(updatedVentas.filter(v => v.fecha === fechaConsulta));
        } catch (err) {
            console.error(err);
            setError('Error al actualizar la venta');
        }
    };

    // Cancela la edición.
    const handleCancel = () => setEditingId(null);

    // Maneja los cambios en los campos de edición.
    const handleFieldChange = (e) => {
        const { name, value } = e.target;
        setEditedData(prev => ({ ...prev, [name]: value }));
    };

    // Elimina una venta individual.
    const handleDelete = async (id) => {
        if (await confirmar('¿Estás seguro de eliminar esta venta?')) {
            try {
                await deleteVenta(id);
                const updatedVentas = ventasOriginales.filter(v => v.id !== id);
                setVentasOriginales(updatedVentas);
                setVentas(updatedVentas.filter(v => v.fecha === fechaConsulta));
            } catch (err) {
                console.error(err);
                setError('Error al eliminar la venta');
            }
        }
    };

    // Maneja la selección de una o varias ventas.
    const handleSelect = (id) => {
        setSelectedIds(prev => prev.includes(id) ? prev.filter(i => i !== id) : [...prev, id]);
    };

    // Selecciona o deselecciona todas las ventas.
    const handleSelectAll = (e) => {
        if (e.target.checked) {
            setSelectedIds(ventas.map(v => v.id));
        } else {
            setSelectedIds([]);
        }
    };

    // Elimina todas las ventas seleccionadas.
    const handleDeleteSelected = async () => {
        if (await confirmar(`¿Estás seguro de eliminar ${selectedIds.length} ventas seleccionadas?`)) {
            try {
                await deleteVentas(selectedIds);
                const updatedVentas = ventasOriginales.filter(v => !selectedIds.includes(v.id));
                setVentasOriginales(updatedVentas);
                setVentas(updatedVentas.filter(v => v.fecha === fechaConsulta));
                setSelectedIds([]);
            } catch (err) {
                console.error(err);
                setError('Error al eliminar las ventas seleccionadas');
            }
        }
    };

    if (loading) return <Spinner animation="border" />;
    if (error && ventasOriginales.length === 0) {
        return (
            <Alert variant="danger">
                {error}{' '}
                <Button variant="link" className="p-0 align-baseline" onClick={() => setReloadToken(t => t + 1)}>
                    Reintentar
                </Button>
            </Alert>
        );
    }

    return (
        <div>
            {error && <Alert variant="danger" onClose={() => setError('')} dismissible>{error}</Alert>}
            <div className="card mb-3">
                <Row className="align-items-end g-3">
                    <Col md={3}>
                        <Form.Group>
                            <Form.Label className="fw-semibold">Fecha</Form.Label>
                            <Form.Control type="date" value={fechaConsulta} onChange={(e) => setFechaConsulta(e.target.value)} />
                        </Form.Group>
                    </Col>
                    <Col md={5}>
                        <Form.Group>
                            <Form.Label className="fw-semibold">Buscar por Receta</Form.Label>
                            <div className="search-wrap">
                                <span className="search-icon"><SearchIcon size={15} /></span>
                                <Form.Control
                                    type="text"
                                    placeholder="Nombre de la receta..."
                                    value={filtroNombre}
                                    onChange={(e) => setFiltroNombre(e.target.value)}
                                />
                            </div>
                        </Form.Group>
                    </Col>
                    <Col md={4} className="d-flex justify-content-end">
                        {selectedIds.length > 0 && (
                            <Button variant="danger" className="d-inline-flex align-items-center gap-2" onClick={handleDeleteSelected}>
                                <TrashIcon size={15} /> Eliminar ({selectedIds.length}) seleccionadas
                            </Button>
                        )}
                    </Col>
                </Row>
            </div>
            <div className="list-card">
                <table className="list-table">
                    <thead>
                        <tr>
                            <th style={{ width: '40px' }}>
                                <Form.Check
                                    type="checkbox"
                                    aria-label="Seleccionar todas las ventas"
                                    onChange={handleSelectAll}
                                    checked={selectedIds.length === ventas.length && ventas.length > 0}
                                />
                            </th>
                            <th>Receta</th>
                            <th>Cantidad</th>
                            <th>Fecha</th>
                            <th>Acciones</th>
                        </tr>
                    </thead>
                    <tbody>
                        {ventas.length === 0 ? (
                            <tr><td colSpan="5" style={{ textAlign: 'center' }}>No hay ventas para la fecha seleccionada.</td></tr>
                        ) : (
                            ventas.map(venta => (
                                <tr
                                    key={venta.id}
                                    style={
                                        editingId === venta.id
                                            ? { backgroundColor: 'var(--warning-soft)' }
                                            : selectedIds.includes(venta.id)
                                                ? { backgroundColor: 'var(--color-primary-soft)' }
                                                : undefined
                                    }
                                >
                                    <td data-label="">
                                        <Form.Check
                                            type="checkbox"
                                            aria-label={`Seleccionar venta de ${venta.receta_nombre} del ${venta.fecha}`}
                                            checked={selectedIds.includes(venta.id)}
                                            onChange={() => handleSelect(venta.id)}
                                        />
                                    </td>
                                    <td data-label="Receta" style={{ fontWeight: 600 }}>{venta.receta_nombre}</td>
                                    {editingId === venta.id ? (
                                        <>
                                            <td data-label="Cantidad"><Form.Control type="number" name="cantidad" value={editedData.cantidad} onChange={handleFieldChange} /></td>
                                            <td data-label="Fecha"><Form.Control type="date" name="fecha" value={editedData.fecha} onChange={handleFieldChange} /></td>
                                            <td data-label="">
                                                <Button variant="primary" size="sm" className="me-2 d-inline-flex align-items-center gap-1" onClick={() => handleSave(venta.id)}>
                                                    <CheckIcon size={13} /> Guardar
                                                </Button>
                                                <Button variant="outline-secondary" size="sm" onClick={handleCancel}>Cancelar</Button>
                                            </td>
                                        </>
                                    ) : (
                                        <>
                                            <td data-label="Cantidad">{venta.cantidad}</td>
                                            <td data-label="Fecha">{venta.fecha}</td>
                                            <td data-label="">
                                                <Button variant="outline-secondary" size="sm" className="me-2 d-inline-flex align-items-center gap-1" onClick={() => handleEdit(venta)}>
                                                    <PencilIcon size={13} /> Editar
                                                </Button>
                                                <Button variant="outline-danger" size="sm" className="d-inline-flex align-items-center gap-1" onClick={() => handleDelete(venta.id)}>
                                                    <TrashIcon size={13} /> Eliminar
                                                </Button>
                                            </td>
                                        </>
                                    )}
                                </tr>
                            ))
                        )}
                    </tbody>
                </table>
            </div>
        </div>
    );
};

export default VentaList;
