import React, { useEffect, useMemo, useState } from 'react';
import { Modal, Button, Form } from 'react-bootstrap';
import { DownloadIcon, iconoDeArea } from '../../components/icons';

const TIPOS_RESUMEN = [
    { clave: 'faltantes', etiqueta: 'Faltante', badge: 'ipv-badge-neg' },
    { clave: 'sobrantes', etiqueta: 'Sobrante', badge: 'ipv-badge-pos' },
    { clave: 'mermas', etiqueta: 'Merma', badge: 'ipv-badge-warn' },
];

// Vista previa del informe antes de exportarlo a PDF: muestra el resumen
// (faltantes/sobrantes/mermas) y las notas generadas automáticamente, y deja
// destildar las que no sean relevantes (p. ej. una diferencia de 15 g, o una
// nota de "cierre anterior" que no interesa) antes de generar el PDF final.
function ReportePreviewModal({ show, onHide, reporteData, onExportar }) {
    const [excluidos, setExcluidos] = useState(() => new Set());

    // Cada vez que se abre con datos nuevos, todo arranca incluido.
    useEffect(() => {
        if (show) setExcluidos(new Set());
    }, [show, reporteData]);

    const filasResumen = useMemo(() => {
        if (!reporteData) return [];
        const filas = [];
        for (const areaNombre in reporteData.resumen) {
            const r = reporteData.resumen[areaNombre];
            TIPOS_RESUMEN.forEach(({ clave, etiqueta, badge }) => {
                (r[clave] || []).forEach((item, idx) => {
                    filas.push({
                        key: `resumen|${areaNombre}|${clave}|${idx}`,
                        areaNombre, etiqueta, badge,
                        texto: `${item.producto}: ${item.cantidad} ${item.um}`,
                    });
                });
            });
        }
        return filas;
    }, [reporteData]);

    const filasNotas = useMemo(() => {
        if (!reporteData) return [];
        const filas = [];
        for (const areaNombre in reporteData.notas) {
            (reporteData.notas[areaNombre] || []).forEach((texto, idx) => {
                filas.push({ key: `nota|${areaNombre}|${idx}`, areaNombre, texto });
            });
        }
        return filas;
    }, [reporteData]);

    const agruparPorArea = (filas) => {
        const porArea = new Map();
        filas.forEach(f => {
            if (!porArea.has(f.areaNombre)) porArea.set(f.areaNombre, []);
            porArea.get(f.areaNombre).push(f);
        });
        return porArea;
    };

    const toggle = (key) => {
        setExcluidos(prev => {
            const next = new Set(prev);
            if (next.has(key)) next.delete(key); else next.add(key);
            return next;
        });
    };

    const totalFilas = filasResumen.length + filasNotas.length;
    const totalExcluidas = excluidos.size;

    const marcarTodo = (incluir) => {
        if (incluir) {
            setExcluidos(new Set());
        } else {
            setExcluidos(new Set([...filasResumen, ...filasNotas].map(f => f.key)));
        }
    };

    const handleExportar = () => {
        if (!reporteData) return;
        // Reconstruye resumen/notas aplicando las exclusiones; las tablas por
        // área (los números reales) se exportan siempre completas, tal cual.
        const resumenFiltrado = {};
        for (const areaNombre in reporteData.resumen) {
            const r = reporteData.resumen[areaNombre];
            resumenFiltrado[areaNombre] = {};
            TIPOS_RESUMEN.forEach(({ clave }) => {
                resumenFiltrado[areaNombre][clave] = (r[clave] || []).filter(
                    (_, idx) => !excluidos.has(`resumen|${areaNombre}|${clave}|${idx}`)
                );
            });
        }
        const notasFiltradas = {};
        for (const areaNombre in reporteData.notas) {
            notasFiltradas[areaNombre] = (reporteData.notas[areaNombre] || []).filter(
                (_, idx) => !excluidos.has(`nota|${areaNombre}|${idx}`)
            );
        }
        onExportar({ ...reporteData, resumen: resumenFiltrado, notas: notasFiltradas });
    };

    const renderGrupo = (titulo, filas) => {
        const porArea = agruparPorArea(filas);
        if (porArea.size === 0) return null;
        return (
            <div className="mb-4">
                <h3 style={{ fontSize: '0.9375rem', fontWeight: 700, marginBottom: '10px' }}>{titulo}</h3>
                {[...porArea.entries()].map(([areaNombre, items]) => {
                    const AreaIcon = iconoDeArea(areaNombre);
                    return (
                        <div key={areaNombre} className="list-card mb-3">
                            <div className="d-flex align-items-center gap-2 px-3 py-2 border-bottom" style={{ fontWeight: 600, fontSize: '0.875rem' }}>
                                <AreaIcon size={15} /> {areaNombre}
                            </div>
                            {items.map(f => (
                                <label key={f.key} className="list-row" style={{ cursor: 'pointer', gap: '10px' }}>
                                    <span className="d-flex align-items-center gap-2" style={{ flex: 1 }}>
                                        <Form.Check
                                            type="checkbox"
                                            checked={!excluidos.has(f.key)}
                                            onChange={() => toggle(f.key)}
                                            className="mb-0"
                                        />
                                        {f.etiqueta && <span className={`ipv-badge ${f.badge}`}>{f.etiqueta}</span>}
                                        <span style={{ color: excluidos.has(f.key) ? 'var(--text-secondary)' : 'var(--text-main)', textDecoration: excluidos.has(f.key) ? 'line-through' : 'none' }}>
                                            {f.texto}
                                        </span>
                                    </span>
                                </label>
                            ))}
                        </div>
                    );
                })}
            </div>
        );
    };

    return (
        <Modal show={show} onHide={onHide} size="lg" centered scrollable>
            <Modal.Header closeButton>
                <Modal.Title>Vista previa del informe</Modal.Title>
            </Modal.Header>
            <Modal.Body>
                {totalFilas === 0 ? (
                    <p style={{ color: 'var(--text-secondary)' }}>
                        No hay diferencias, mermas ni notas que revisar: el informe solo tendrá las tablas con los
                        valores por área.
                    </p>
                ) : (
                    <>
                        <div className="d-flex align-items-center justify-content-between mb-3">
                            <p className="mb-0" style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
                                Destilda lo que no quieras incluir en el PDF. Las tablas por área se exportan siempre
                                completas.
                            </p>
                            <div className="d-flex gap-2 flex-shrink-0">
                                <Button variant="outline-secondary" size="sm" onClick={() => marcarTodo(true)}>Marcar todo</Button>
                                <Button variant="outline-secondary" size="sm" onClick={() => marcarTodo(false)}>Desmarcar todo</Button>
                            </div>
                        </div>
                        {renderGrupo('Resumen por área', filasResumen)}
                        {renderGrupo('Notas y comentarios', filasNotas)}
                    </>
                )}
            </Modal.Body>
            <Modal.Footer>
                {totalExcluidas > 0 && (
                    <span className="me-auto" style={{ fontSize: '0.8125rem', color: 'var(--text-secondary)' }}>
                        {totalExcluidas} de {totalFilas} excluidas
                    </span>
                )}
                <Button variant="outline-secondary" onClick={onHide}>Cancelar</Button>
                <Button variant="primary" className="d-inline-flex align-items-center gap-2" onClick={handleExportar}>
                    <DownloadIcon size={15} /> Generar PDF
                </Button>
            </Modal.Footer>
        </Modal>
    );
}

export default ReportePreviewModal;
