import React from 'react';
import EditableCell from './EditableCell';
import { LockIcon } from '../../components/icons';

// Renderiza la tabla de inventario de un área: columna de ingrediente fija,
// celdas editables con su input, y celdas calculadas (Consumo, Final Consumo)
// distinguidas visualmente (fondo tintado + candado) para que no se confundan
// con las editables.
function AreaIPVTab({ areaData, onItemChange, onCommentChange }) {
    const badgeClass = (diferencia) => {
        if (diferencia < 0) return 'ipv-badge ipv-badge-neg';
        if (diferencia > 0) return 'ipv-badge ipv-badge-pos';
        return 'ipv-badge ipv-badge-zero';
    };

    return (
        <div className="ipv-table-wrap">
            <table className="ipv-table">
                <thead>
                    <tr>
                        <th className="ipv-col-ing" style={{ width: '20%' }}>Ingrediente</th>
                        <th style={{ width: '9%' }}>Inicio</th>
                        <th style={{ width: '9%' }}>Entradas</th>
                        <th style={{ width: '10%' }}>Consumo</th>
                        <th style={{ width: '8%' }}>Merma</th>
                        <th style={{ width: '11%' }}>Otras Salidas</th>
                        <th style={{ width: '10%' }}>Final Físico</th>
                        <th style={{ width: '10%' }}>Final Consumo</th>
                        <th style={{ width: '13%' }}>Diferencia</th>
                    </tr>
                </thead>
                <tbody>
                    {areaData.map(item => (
                        <tr key={item.producto_id}>
                            <td className="ipv-col-ing">{item.producto_nombre || 'Producto no encontrado'}</td>
                            <td>
                                <EditableCell
                                    label={`Inicio de ${item.producto_nombre}`}
                                    value={item.inicio}
                                    onChange={(value) => onItemChange(item.producto_id, 'inicio', value)}
                                    onCommentChange={(comment) => onCommentChange(item.producto_id, 'inicio', comment)}
                                    comment={item.comentarios ? item.comentarios.inicio : ''}
                                />
                            </td>
                            <td>
                                <EditableCell
                                    label={`Entradas de ${item.producto_nombre}`}
                                    value={item.entradas}
                                    onChange={(value) => onItemChange(item.producto_id, 'entradas', value)}
                                    onCommentChange={(comment) => onCommentChange(item.producto_id, 'entradas', comment)}
                                    comment={item.comentarios ? item.comentarios.entradas : ''}
                                />
                            </td>
                            <td className="ipv-cell-readonly">
                                <span className="ipv-cell-readonly-inner">
                                    <LockIcon size={12} style={{ opacity: 0.6 }} />
                                    {(item.consumo || 0).toFixed(3)}
                                </span>
                            </td>
                            <td>
                                <EditableCell
                                    label={`Merma de ${item.producto_nombre}`}
                                    value={item.merma}
                                    onChange={(value) => onItemChange(item.producto_id, 'merma', value)}
                                    onCommentChange={(comment) => onCommentChange(item.producto_id, 'merma', comment)}
                                    comment={item.comentarios ? item.comentarios.merma : ''}
                                />
                            </td>
                            <td>
                                <EditableCell
                                    label={`Otras salidas de ${item.producto_nombre}`}
                                    value={item.otras_salidas}
                                    onChange={(value) => onItemChange(item.producto_id, 'otras_salidas', value)}
                                    onCommentChange={(comment) => onCommentChange(item.producto_id, 'otras_salidas', comment)}
                                    comment={item.comentarios ? item.comentarios.otras_salidas : ''}
                                />
                            </td>
                            <td>
                                <EditableCell
                                    label={`Final físico de ${item.producto_nombre}`}
                                    value={item.final_fisico}
                                    onChange={(value) => onItemChange(item.producto_id, 'final_fisico', value)}
                                    onCommentChange={(comment) => onCommentChange(item.producto_id, 'final_fisico', comment)}
                                    comment={item.comentarios ? item.comentarios.final_fisico : ''}
                                />
                            </td>
                            <td className="ipv-cell-readonly">
                                <span className="ipv-cell-readonly-inner">
                                    <LockIcon size={12} style={{ opacity: 0.6 }} />
                                    {Number(item.final_teorico || 0).toFixed(3)}
                                </span>
                            </td>
                            <td style={{ textAlign: 'right' }}>
                                <span className={badgeClass(item.diferencia)}>
                                    {Number(item.diferencia || 0).toFixed(3)}
                                </span>
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    );
}

export default AreaIPVTab;
