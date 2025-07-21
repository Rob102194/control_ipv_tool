import React, { useState, useEffect } from 'react';
import { Table, Button, Container, Alert, Spinner, Form, Row, Col } from 'react-bootstrap';
import { getVentas, updateVenta, deleteVenta, importVentas, deleteVentas } from '../../api/ventaApi';

// Componente principal para la gestión de ventas.
const VentaList = () => {
    // Estado para gestionar la vista actual ('importar' o 'consultar').
    const [view, setView] = useState('importar');

    // Renderiza el componente principal y redirige a la vista seleccionada.
    return (
        <Container className="mt-4">
            <h1>Gestión de Ventas</h1>
            <hr />
            <Button variant="primary" className="me-2" onClick={() => setView('importar')}>Importar Ventas</Button>
            <Button variant="info" onClick={() => setView('consultar')}>Consultar Ventas</Button>
            <hr />
            {view === 'importar' ? <ImportarVentas /> : <ConsultarVentas />}
        </Container>
    );
};

// Componente para la importación de ventas.
const ImportarVentas = () => {
    const [file, setFile] = useState(null);
    const [fecha, setFecha] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    // Maneja la importación del archivo de ventas.
    const handleImport = async () => {
        if (!file || !fecha) {
            alert("Por favor, seleccione un archivo y una fecha.");
            return;
        }
        setLoading(true);
        setError('');
        try {
            await importVentas(file, fecha);
            alert('¡Ventas importadas con éxito!');
            setFile(null);
            setFecha('');
        } catch (err) {
            setError('Error al importar las ventas');
            alert(err.response?.data?.error || "Error al importar ventas");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div>
            <h2>Importar Ventas desde Excel</h2>
            {error && <Alert variant="danger">{error}</Alert>}
            <Row>
                <Col md={6}>
                    <Form.Group>
                        <Form.Label>Archivo Excel</Form.Label>
                        <Form.Control type="file" accept=".xlsx, .xls" onChange={(e) => setFile(e.target.files[0])} />
                    </Form.Group>
                </Col>
                <Col md={6}>
                    <Form.Group>
                        <Form.Label>Fecha de las Ventas</Form.Label>
                        <Form.Control type="date" value={fecha} onChange={(e) => setFecha(e.target.value)} />
                    </Form.Group>
                </Col>
            </Row>
            <Button variant="secondary" onClick={handleImport} className="mt-3" disabled={!file || !fecha || loading}>
                {loading ? <><Spinner as="span" animation="border" size="sm" /> Importando...</> : 'Importar'}
            </Button>
        </div>
    );
};

// Componente para consultar, editar y eliminar ventas.
const ConsultarVentas = () => {
    const [ventasOriginales, setVentasOriginales] = useState([]);
    const [ventas, setVentas] = useState([]);
    const [filtroNombre, setFiltroNombre] = useState('');
    const [fechaConsulta, setFechaConsulta] = useState(new Date().toISOString().split('T')[0]);
    const [editingId, setEditingId] = useState(null);
    const [editedData, setEditedData] = useState({});
    const [selectedIds, setSelectedIds] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState('');

    // Carga las ventas al montar el componente.
    useEffect(() => {
        const fetchVentas = async () => {
            setLoading(true);
            try {
                const response = await getVentas();
                setVentasOriginales(response.data);
            } catch (err) {
                setError('Error al cargar las ventas');
            } finally {
                setLoading(false);
            }
        };
        fetchVentas();
    }, []);

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
        if (window.confirm('¿Estás seguro de eliminar esta venta?')) {
            try {
                await deleteVenta(id);
                const updatedVentas = ventasOriginales.filter(v => v.id !== id);
                setVentasOriginales(updatedVentas);
                setVentas(updatedVentas.filter(v => v.fecha === fechaConsulta));
            } catch (err) {
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
        if (window.confirm(`¿Estás seguro de eliminar ${selectedIds.length} ventas seleccionadas?`)) {
            try {
                await deleteVentas(selectedIds);
                const updatedVentas = ventasOriginales.filter(v => !selectedIds.includes(v.id));
                setVentasOriginales(updatedVentas);
                setVentas(updatedVentas.filter(v => v.fecha === fechaConsulta));
                setSelectedIds([]);
            } catch (err) {
                setError('Error al eliminar las ventas seleccionadas');
            }
        }
    };

    if (loading) return <Spinner animation="border" />;
    if (error) return <Alert variant="danger">{error}</Alert>;

    return (
        <div>
            <h2>Consultar Ventas</h2>
            <Row className="mb-3">
                <Col md={4}>
                    <Form.Group>
                        <Form.Label>Fecha</Form.Label>
                        <Form.Control type="date" value={fechaConsulta} onChange={(e) => setFechaConsulta(e.target.value)} />
                    </Form.Group>
                </Col>
                <Col md={4}>
                    <Form.Group>
                        <Form.Label>Buscar por Receta</Form.Label>
                        <Form.Control
                            type="text"
                            placeholder="Nombre de la receta..."
                            value={filtroNombre}
                            onChange={(e) => setFiltroNombre(e.target.value)}
                        />
                    </Form.Group>
                </Col>
                <Col md={4} className="d-flex align-items-end">
                    {selectedIds.length > 0 && (
                        <Button variant="danger" onClick={handleDeleteSelected}>
                            Eliminar ({selectedIds.length}) Seleccionadas
                        </Button>
                    )}
                </Col>
            </Row>
            <Table striped bordered hover responsive>
                <thead>
                    <tr>
                        <th>
                            <Form.Check 
                                type="checkbox"
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
                        <tr><td colSpan="5" className="text-center">No hay ventas para la fecha seleccionada.</td></tr>
                    ) : (
                        ventas.map(venta => (
                            <tr key={venta.id}>
                                <td>
                                    <Form.Check 
                                        type="checkbox"
                                        checked={selectedIds.includes(venta.id)}
                                        onChange={() => handleSelect(venta.id)}
                                    />
                                </td>
                                <td>{venta.receta_nombre}</td>
                                {editingId === venta.id ? (
                                    <>
                                        <td><Form.Control type="number" name="cantidad" value={editedData.cantidad} onChange={handleFieldChange} /></td>
                                        <td><Form.Control type="date" name="fecha" value={editedData.fecha} onChange={handleFieldChange} /></td>
                                        <td>
                                            <Button variant="success" size="sm" onClick={() => handleSave(venta.id)}>Guardar</Button>
                                            <Button variant="secondary" size="sm" className="ms-2" onClick={handleCancel}>Cancelar</Button>
                                        </td>
                                    </>
                                ) : (
                                    <>
                                        <td>{venta.cantidad}</td>
                                        <td>{venta.fecha}</td>
                                        <td>
                                            <Button variant="warning" size="sm" onClick={() => handleEdit(venta)}>Editar</Button>
                                            <Button variant="danger" size="sm" className="ms-2" onClick={() => handleDelete(venta.id)}>Eliminar</Button>
                                        </td>
                                    </>
                                )}
                            </tr>
                        ))
                    )}
                </tbody>
            </Table>
        </div>
    );
};

export default VentaList;
