import React, { useState, useEffect } from 'react';
import { Form, Button, Container, Row, Col, Modal, Spinner, Alert } from 'react-bootstrap';
import ipvApi from '../../api/ipvApi';
import { obtenerProductos } from '../../api/productoApi';
import AreaIPVTab from './AreaIPVTab';
import ModeloIPV from './ModeloIPV';
import IPVRegistrosList from './IPVRegistrosList';
import ReporteIPV from './ReporteIPV';

// Componente principal para el control del Inventario Diario (IPV).
function IPVControl() {
    const [fecha, setFecha] = useState('');
    const [inventario, setInventario] = useState({});
    const [productos, setProductos] = useState([]);
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');
    const [showRegistroModal, setShowRegistroModal] = useState(false);
    const [showModeloModal, setShowModeloModal] = useState(false);
    const [showRegistrosListModal, setShowRegistrosListModal] = useState(false);
    const [inventarioAnterior, setInventarioAnterior] = useState(null);

    useEffect(() => {
        const fetchProductos = async () => {
            try {
                const response = await obtenerProductos();
                setProductos(response.data);
            } catch (error) {
                console.error("Error al cargar productos", error);
                setError("No se pudieron cargar los productos, la UM no aparecerá en el reporte.");
            }
        };
        fetchProductos();
    }, []);

    const handleSelectRegistro = (fechaSeleccionada) => {
        setFecha(fechaSeleccionada);
        setShowRegistrosListModal(false);
        handleCargarDatos(fechaSeleccionada);
    };

    const buildReportDataFromState = () => {
        const reporte = {
            fecha: fecha,
            areas: {},
            resumen: {},
            notas: {}
        };

        const productosMap = new Map(productos.map(p => [p.id, p]));

        for (const areaNombre in inventario) {
            if (!reporte.areas[areaNombre]) reporte.areas[areaNombre] = [];
            if (!reporte.resumen[areaNombre]) {
                reporte.resumen[areaNombre] = { faltantes: [], sobrantes: [], mermas: [] };
            }

            inventario[areaNombre].forEach(item => {
                const producto = productosMap.get(item.producto_id);
                const um = producto ? producto.unidad_medida : 'N/A';

                if (item.comentarios) {
                    if (!reporte.notas[areaNombre]) {
                        reporte.notas[areaNombre] = [];
                    }
                    for (const [campo, texto] of Object.entries(item.comentarios)) {
                        if (texto && texto.trim()) {
                            const cantidad = item[campo] || 0;
                            const nota = `${item.producto_nombre} ${cantidad} ${um} ${campo}: ${texto}`;
                            reporte.notas[areaNombre].push(nota);
                        }
                    }
                }

                if (inventarioAnterior && inventarioAnterior[areaNombre]) {
                    const itemAnterior = inventarioAnterior[areaNombre].find(p => p.producto_id === item.producto_id);
                    if (itemAnterior && item.inicio !== itemAnterior.final_fisico) {
                        if (!reporte.notas[areaNombre]) {
                            reporte.notas[areaNombre] = [];
                        }
                        const diffNota = `Diferencia con cierre anterior para ${item.producto_nombre}: Inicio: ${item.inicio}, Cierre anterior: ${itemAnterior.final_fisico}`;
                        reporte.notas[areaNombre].push(diffNota);
                    }
                }

                reporte.areas[areaNombre].push({
                    producto: item.producto_nombre,
                    um: um,
                    inicio: item.inicio,
                    entradas: item.entradas,
                    consumo: item.consumo,
                    merma: item.merma,
                    otras_salidas: item.otras_salidas,
                    final_teorico: item.final_teorico,
                    final_fisico: item.final_fisico,
                    diferencia: item.diferencia
                });

                if (item.diferencia < 0) {
                    reporte.resumen[areaNombre].faltantes.push(`${item.producto_nombre}: ${Math.abs(item.diferencia)} ${um}`);
                } else if (item.diferencia > 0) {
                    reporte.resumen[areaNombre].sobrantes.push(`${item.producto_nombre}: ${item.diferencia} ${um}`);
                }
                if (item.merma > 0) {
                    reporte.resumen[areaNombre].mermas.push(`${item.producto_nombre}: ${item.merma} ${um}`);
                }
            });
        }
        return reporte;
    };

    const handleGenerarReporte = () => {
        if (!fecha) {
            setError('Por favor, seleccione una fecha para generar el reporte.');
            return;
        }
        const reporteData = buildReportDataFromState();
        ReporteIPV(reporteData);
    };
    // Maneja la carga inicial de datos al seleccionar una fecha.
    const handleCargarDatos = async (fechaACargar) => {
        const fechaParaCargar = fechaACargar || fecha;
        if (!fechaParaCargar) {
            setError('Por favor, seleccione una fecha.');
            return;
        }
        setLoading(true);
        setError('');

        const fechaParts = fechaParaCargar.split('-');
        const fechaObj = new Date(fechaParts[0], fechaParts[1] - 1, fechaParts[2]);
        fechaObj.setDate(fechaObj.getDate() - 1);
        const fechaAnteriorString = fechaObj.toISOString().split('T')[0];

        try {
            const [response, responseAnterior] = await Promise.all([
                ipvApi.getEstado(fechaParaCargar),
                ipvApi.getEstado(fechaAnteriorString).catch(e => { console.error(e); return null; })
            ]);

            const inventarioConComentariosParseados = response.data;
            for (const areaNombre in inventarioConComentariosParseados) {
                inventarioConComentariosParseados[areaNombre] = inventarioConComentariosParseados[areaNombre].map(item => {
                    try {
                        const comentarios = item.comentario ? JSON.parse(item.comentario) : {};
                        return { ...item, comentarios };
                    } catch (e) {
                        return { ...item, comentarios: {} };
                    }
                });
            }
            setInventario(inventarioConComentariosParseados);

            if (responseAnterior) {
                setInventarioAnterior(responseAnterior.data);
            } else {
                setInventarioAnterior(null);
            }

            setShowRegistroModal(true);
        } catch (err) {
            setError('Error al cargar los datos del inventario.');
            console.error(err);
        } finally {
            setLoading(false);
        }
    };

    // Maneja el cálculo del consumo.
    const handleCalcularConsumo = async () => {
        if (!fecha) {
            setError('Por favor, seleccione una fecha antes de calcular el consumo.');
            return;
        }
        setLoading(true);
        try {
            const response = await ipvApi.calcularConsumo(fecha);
            const consumos = response.data;

            if (Object.keys(consumos).length === 0) {
                alert('No se encontraron ventas para la fecha seleccionada. El consumo se mantendrá en cero.');
            }
            
            // Actualiza el estado del inventario con los consumos calculados.
            const nuevoInventario = { ...inventario };
            Object.keys(consumos).forEach(key => {
                const [productoId, areaId] = key.split('|');
                for (const areaNombre in nuevoInventario) {
                    const area = nuevoInventario[areaNombre];
                    const itemIndex = area.findIndex(item => item.producto_id === productoId && item.area_id === areaId);
                    if (itemIndex > -1) {
                        area[itemIndex].consumo = consumos[key];
                    }
                }
            });
            setInventario(nuevoInventario);
        } catch (err) {
            setError('Error al calcular el consumo.');
        } finally {
            setLoading(false);
        }
    };

    // Maneja el cálculo de diferencias.
    const handleCalcularDiferencias = () => {
        const nuevoInventario = { ...inventario };
        for (const areaNombre in nuevoInventario) {
            nuevoInventario[areaNombre] = nuevoInventario[areaNombre].map(item => {
                if (item.final_fisico === null || item.final_fisico === undefined) {
                    return item; // No calcula si el final físico no está ingresado.
                }
                const final_teorico = (item.inicio + item.entradas) - item.consumo - item.merma - item.otras_salidas;
                const diferencia = item.final_fisico - final_teorico;
                return { ...item, final_teorico, diferencia };
            });
        }
        setInventario(nuevoInventario);
    };

    // Limpia los datos del formulario.
    const handleLimpiarDatos = () => {
        const inventarioLimpio = { ...inventario };
        for (const areaNombre in inventarioLimpio) {
            inventarioLimpio[areaNombre] = inventarioLimpio[areaNombre].map(item => ({
                ...item,
                inicio: 0,
                entradas: 0,
                consumo: 0,
                merma: 0,
                otras_salidas: 0,
                final_fisico: 0,
                final_teorico: 0,
                diferencia: 0,
                comentarios: {}
            }));
        }
        setInventario(inventarioLimpio);
    };

    // Maneja el guardado del registro completo.
    const handleGuardar = async () => {
        setLoading(true);
        try {
            const dataToSave = Object.values(inventario).flat().map(item => {
                const newItem = { ...item };
                if (newItem.comentarios) {
                    newItem.comentario = JSON.stringify(newItem.comentarios);
                }
                return newItem;
            });

            await ipvApi.guardar(dataToSave);
            alert('¡Registro de inventario guardado con éxito!');
        } catch (err) {
            setError('Error al guardar el registro.');
        } finally {
            setLoading(false);
        }
    };

    // Maneja los cambios en los campos de la tabla.
    const handleItemChange = (areaNombre, productoId, field, value) => {
        setInventario(prevInventario => {
            const nuevoInventario = JSON.parse(JSON.stringify(prevInventario));
            const area = nuevoInventario[areaNombre];
            const itemIndex = area.findIndex(item => item.producto_id === productoId);
            if (itemIndex > -1) {
                const numericValue = Math.max(0, parseFloat(value) || 0);
                area[itemIndex][field] = numericValue;
            }
            return nuevoInventario;
        });
    };

    const handleCommentChange = (areaNombre, productoId, field, comment) => {
        setInventario(prevInventario => {
            const nuevoInventario = JSON.parse(JSON.stringify(prevInventario));
            const area = nuevoInventario[areaNombre];
            const itemIndex = area.findIndex(item => item.producto_id === productoId);
            if (itemIndex > -1) {
                if (!area[itemIndex].comentarios) {
                    area[itemIndex].comentarios = {};
                }
                area[itemIndex].comentarios[field] = comment;
            }
            return nuevoInventario;
        });
    };

    return (
        <Container fluid>
            <h2 className="my-4">Control de Inventario Diario (IPV)</h2>
            
            <Form.Group as={Row} className="mb-3" controlId="fechaSelector">
                <Form.Label column sm={2}>
                    Seleccione la Fecha
                </Form.Label>
                <Col sm={3}>
                    <Form.Control
                        type="date"
                        value={fecha}
                        onChange={(e) => setFecha(e.target.value)}
                    />
                </Col>
            </Form.Group>

            <Button variant="primary" onClick={() => setShowModeloModal(true)}>
                Crear Modelos
            </Button>{' '}
            <Button variant="secondary" onClick={() => handleCargarDatos()} disabled={!fecha || loading}>
                {loading ? <Spinner as="span" animation="border" size="sm" /> : 'Nuevo Registro Diario'}
            </Button>{' '}
            <Button variant="outline-primary" onClick={() => setShowRegistrosListModal(true)}>
                Consultar Registros
            </Button>

            {/* Modal para el registro diario */}
            <Modal show={showRegistroModal} onHide={() => setShowRegistroModal(false)} size="xl" centered>
                <Modal.Header closeButton>
                    <Modal.Title>Nuevo Registro de IPV para el {fecha}</Modal.Title>
                </Modal.Header>
                <Modal.Body style={{ overflowX: 'auto' }}>
                    {error && <Alert variant="danger">{error}</Alert>}
                    
                    {Object.keys(inventario).map(areaNombre => (
                        <div key={areaNombre}>
                            <h3>{areaNombre}</h3>
                            <AreaIPVTab
                                areaData={inventario[areaNombre]}
                                onItemChange={(productoId, field, value) => handleItemChange(areaNombre, productoId, field, value)}
                                onCommentChange={(productoId, field, comment) => handleCommentChange(areaNombre, productoId, field, comment)}
                            />
                        </div>
                    ))}
                </Modal.Body>
                <Modal.Footer>
                    <Button variant="primary" onClick={handleCalcularConsumo} disabled={loading}>Calcular Consumo</Button>
                    <Button variant="info" onClick={handleCalcularDiferencias} disabled={loading}>Calcular Diferencias</Button>
                    <Button variant="warning" onClick={handleLimpiarDatos} disabled={loading}>Limpiar Datos</Button>
                    <Button variant="success" onClick={handleGuardar} disabled={loading}>Guardar Registro</Button>
                    <Button variant="light" onClick={handleGenerarReporte} disabled={loading}>Generar Reporte</Button>
                    <Button variant="outline-secondary" onClick={() => setShowRegistroModal(false)}>Cerrar</Button>
                </Modal.Footer>
            </Modal>

            {/* Modal para la configuración de modelos */}
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

            {/* Modal para la lista de registros */}
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
        </Container>
    );
}

export default IPVControl;
