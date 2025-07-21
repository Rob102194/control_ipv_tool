import React from 'react';
import { Modal, Button, Spinner, Alert } from 'react-bootstrap';
import AreaIPVTab from './AreaIPVTab';
import ReporteIPV from './ReporteIPV';

function RegistroDiarioModal({
    show,
    onHide,
    fecha,
    inventario,
    loading,
    error,
    handleCalcularConsumo,
    handleCalcularDiferencias,
    handleLimpiarDatos,
    handleGuardar,
    handleItemChange,
    handleCommentChange,
    buildReportData
}) {

    const handleGenerarReporte = () => {
        if (!fecha) {
            alert('Por favor, seleccione una fecha para generar el reporte.');
            return;
        }
        const reporteData = buildReportData();
        ReporteIPV(reporteData);
    };

    return (
        <Modal show={show} onHide={onHide} size="xl" centered>
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
                <Button variant="outline-secondary" onClick={onHide}>Cerrar</Button>
            </Modal.Footer>
        </Modal>
    );
}

export default RegistroDiarioModal;
