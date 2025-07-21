import React from 'react';
import { Table } from 'react-bootstrap';
import EditableCell from './EditableCell';

// Componente para renderizar la tabla de inventario de un área específica.
function AreaIPVTab({ areaData, onItemChange, onCommentChange }) {

    // Determina el estilo de la celda de diferencia.
    const getDiferenciaStyle = (diferencia) => {
        if (diferencia < 0) return { backgroundColor: '#8B0000' }; // Rojo oscuro para faltantes
        if (diferencia > 0) return { backgroundColor: '#006400' }; // Verde oscuro para sobrantes
        return {};
    };

    return (
        <Table striped bordered hover size="sm" className="text-center align-middle">
            <thead>
                <tr>
                    <th style={{ width: '200px', minWidth: '200px', textAlign: 'left' }}>Ingrediente</th>
                    <th style={{ width: '125px', minWidth: '125px' }}>Inicio</th>
                    <th style={{ width: '125px', minWidth: '125px' }}>Entradas</th>
                    <th style={{ width: '95px', minWidth: '95px' }}>Consumo</th>
                    <th style={{ width: '125px', minWidth: '125px' }}>Merma</th>
                    <th style={{ width: '125px', minWidth: '125px' }}>Otras Salidas</th>
                    <th style={{ width: '125px', minWidth: '125px' }}>Final Físisco</th>
                    <th style={{ width: '95px', minWidth: '95px' }}>Final Consumo</th>
                    <th style={{ width: '95px', minWidth: '95px' }}>Diferencias</th>
                </tr>
            </thead>
            <tbody>
                {areaData.map(item => (
                    <tr key={item.producto_id}>
                        <td style={{ textAlign: 'left' }}>{item.producto_nombre || 'Producto no encontrado'}</td>
                        <td>
                            <EditableCell
                                value={item.inicio}
                                onChange={(value) => onItemChange(item.producto_id, 'inicio', value)}
                                onCommentChange={(comment) => onCommentChange(item.producto_id, 'inicio', comment)}
                                comment={item.comentarios ? item.comentarios.inicio : ''}
                            />
                        </td>
                        <td>
                            <EditableCell
                                value={item.entradas}
                                onChange={(value) => onItemChange(item.producto_id, 'entradas', value)}
                                onCommentChange={(comment) => onCommentChange(item.producto_id, 'entradas', comment)}
                                comment={item.comentarios ? item.comentarios.entradas : ''}
                            />
                        </td>
                        <td>{(item.consumo || 0).toFixed(3)}</td>
                        <td>
                            <EditableCell
                                value={item.merma}
                                onChange={(value) => onItemChange(item.producto_id, 'merma', value)}
                                onCommentChange={(comment) => onCommentChange(item.producto_id, 'merma', comment)}
                                comment={item.comentarios ? item.comentarios.merma : ''}
                            />
                        </td>
                        <td>
                            <EditableCell
                                value={item.otras_salidas}
                                onChange={(value) => onItemChange(item.producto_id, 'otras_salidas', value)}
                                onCommentChange={(comment) => onCommentChange(item.producto_id, 'otras_salidas', comment)}
                                comment={item.comentarios ? item.comentarios.otras_salidas : ''}
                            />
                        </td>
                        <td>
                            <EditableCell
                                value={item.final_fisico}
                                onChange={(value) => onItemChange(item.producto_id, 'final_fisico', value)}
                                onCommentChange={(comment) => onCommentChange(item.producto_id, 'final_fisico', comment)}
                                comment={item.comentarios ? item.comentarios.final_fisico : ''}
                            />
                        </td>
                        <td>{Number(item.final_teorico || 0).toFixed(3)}</td>
                        <td style={getDiferenciaStyle(item.diferencia)}>
                            {Number(item.diferencia || 0).toFixed(3)}
                        </td>
                    </tr>
                ))}
            </tbody>
        </Table>
    );
}

export default AreaIPVTab;
