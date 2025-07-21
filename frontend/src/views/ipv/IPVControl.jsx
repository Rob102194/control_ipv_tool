import React, { useState } from 'react';
import { Form, Button, Container, Row, Col, Modal, Spinner } from 'react-bootstrap';
import { useIPV } from '../../hooks/useIPV';
import ModeloIPV from './ModeloIPV';
import IPVRegistrosList from './IPVRegistrosList';
import RegistroDiarioModal from './RegistroDiarioModal';

function IPVControl() {
    const {
        fecha,
        setFecha,
        inventario,
        loading,
        error,
        handleCargarDatos,
        handleCalcularConsumo,
        handleCalcularDiferencias,
        handleLimpiarDatos,
        handleGuardar,
        handleItemChange,
        handleCommentChange,
        buildReportData
    } = useIPV();

    const [showRegistroModal, setShowRegistroModal] = useState(false);
    const [showModeloModal, setShowModeloModal] = useState(false);
    const [showRegistrosListModal, setShowRegistrosListModal] = useState(false);

    const handleNuevoRegistroClick = async () => {
        const data = await handleCargarDatos(fecha);
        if (data) {
            setShowRegistroModal(true);
        }
    };

    const handleSelectRegistro = async (fechaSeleccionada) => {
        setFecha(fechaSeleccionada);
        setShowRegistrosListModal(false);
        const data = await handleCargarDatos(fechaSeleccionada);
        if (data) {
            setShowRegistroModal(true);
        }
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
            <Button variant="secondary" onClick={handleNuevoRegistroClick} disabled={!fecha || loading}>
                {loading ? <Spinner as="span" animation="border" size="sm" /> : 'Nuevo Registro Diario'}
            </Button>{' '}
            <Button variant="outline-primary" onClick={() => setShowRegistrosListModal(true)}>
                Consultar Registros
            </Button>

            <RegistroDiarioModal
                show={showRegistroModal}
                onHide={() => setShowRegistroModal(false)}
                fecha={fecha}
                inventario={inventario}
                loading={loading}
                error={error}
                handleCalcularConsumo={handleCalcularConsumo}
                handleCalcularDiferencias={handleCalcularDiferencias}
                handleLimpiarDatos={handleLimpiarDatos}
                handleGuardar={handleGuardar}
                handleItemChange={handleItemChange}
                handleCommentChange={handleCommentChange}
                buildReportData={buildReportData}
            />

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
        </Container>
    );
}

export default IPVControl;
