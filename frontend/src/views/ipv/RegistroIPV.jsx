import React, { useEffect, useState, useMemo, useRef } from 'react';
import { useParams, useSearchParams, Link } from 'react-router-dom';
import { Spinner, Alert } from 'react-bootstrap';
import { useIPV } from '../../hooks/useIPV';
import AreaIPVTab from './AreaIPVTab';
import ReporteIPV from './ReporteIPV';
import { formatDateEs } from '../../utils/date';
import {
    ArrowLeftIcon,
    DownloadIcon,
    CheckIcon,
    LockIcon,
    ChevronDownIcon,
    iconoDeArea,
} from '../../components/icons';

// Registro diario de IPV como página completa (antes vivía en un modal).
// Cada área se edita en su propia pestaña; el consumo y el final teórico
// son calculados por el servidor y se muestran como solo lectura.
function RegistroIPV() {
    const { fecha } = useParams();
    const [searchParams, setSearchParams] = useSearchParams();

    const {
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
        buildReportData,
    } = useIPV();

    // La fecha "oficial" de esta página es la de la URL (useParams), pero
    // handleCalcularConsumo/buildReportData usan el estado interno `fecha`
    // del hook, que nadie más pone al día aquí (a diferencia de IPVControl):
    // sin este efecto queda vacío y "Calcular Consumo" siempre rechaza la
    // fecha como si no se hubiera seleccionado ninguna.
    useEffect(() => {
        if (fecha) setFecha(fecha);
    }, [fecha, setFecha]);

    const [dirty, setDirty] = useState(false);
    const areaFromUrl = searchParams.get('area');
    const [activeArea, setActiveArea] = useState(areaFromUrl || '');

    // Menú "Más acciones" propio (sin Dropdown de react-bootstrap): el proyecto
    // carga a la vez bootstrap.bundle.min.js y react-bootstrap, y ambos pelean
    // por el estado del Dropdown (se abre y se cierra solo). Ver main.jsx.
    const [showMore, setShowMore] = useState(false);
    const moreMenuRef = useRef(null);
    useEffect(() => {
        if (!showMore) return;
        const onDocClick = (e) => {
            if (moreMenuRef.current && !moreMenuRef.current.contains(e.target)) setShowMore(false);
        };
        document.addEventListener('mousedown', onDocClick);
        return () => document.removeEventListener('mousedown', onDocClick);
    }, [showMore]);

    useEffect(() => {
        if (fecha) handleCargarDatos(fecha);
    }, [fecha, handleCargarDatos]);

    const areas = useMemo(() => Object.keys(inventario), [inventario]);

    useEffect(() => {
        if (areas.length === 0) return;
        if (!activeArea || !areas.includes(activeArea)) {
            setActiveArea(areas[0]);
        }
    }, [areas, activeArea]);

    const seleccionarArea = (areaNombre) => {
        setActiveArea(areaNombre);
        setSearchParams({ area: areaNombre }, { replace: true });
    };

    const onItemChange = (...args) => {
        handleItemChange(...args);
        setDirty(true);
    };

    const onCommentChange = (...args) => {
        handleCommentChange(...args);
        setDirty(true);
    };

    const onGuardar = async () => {
        await handleGuardar();
        setDirty(false);
    };

    const onLimpiar = () => {
        if (window.confirm('Esto borrará los valores introducidos en todas las áreas para esta fecha (sin guardar). ¿Continuar?')) {
            handleLimpiarDatos();
            setDirty(true);
        }
    };

    const onGenerarReporte = () => {
        const reporteData = buildReportData();
        ReporteIPV(reporteData);
    };

    const AreaIcon = iconoDeArea(activeArea);

    return (
        <div>
            <div className="d-flex align-items-center justify-content-between flex-wrap gap-3 py-3 px-4 border-bottom"
                style={{ position: 'sticky', top: 0, zIndex: 10, backgroundColor: 'var(--card-bg)' }}>
                <div className="d-flex align-items-center gap-3">
                    <Link to="/" className="ipv-icon-btn" aria-label="Volver a Control IPV">
                        <ArrowLeftIcon size={18} />
                    </Link>
                    <div>
                        <div className="text-uppercase fw-bold" style={{ fontSize: '0.6875rem', letterSpacing: '0.05em', color: 'var(--text-secondary)' }}>
                            <Link to="/" style={{ color: 'inherit', textDecoration: 'none' }}>Control IPV</Link>
                            {' / '}<span style={{ color: 'var(--color-primary)' }}>Registro diario</span>
                        </div>
                        <h1 className="mb-0" style={{ fontSize: '1.3125rem', fontWeight: 700 }}>{formatDateEs(fecha)}</h1>
                    </div>
                </div>
                <div className="d-flex align-items-center gap-2 flex-wrap">
                    {dirty && (
                        <span className="ipv-status-pill">
                            <span className="ipv-status-dot"></span>
                            Cambios sin guardar
                        </span>
                    )}
                    <div className="position-relative" ref={moreMenuRef}>
                        <button
                            type="button"
                            className="btn btn-outline-secondary d-inline-flex align-items-center gap-2"
                            disabled={loading}
                            onClick={() => setShowMore(s => !s)}
                        >
                            Más acciones <ChevronDownIcon size={14} />
                        </button>
                        {showMore && (
                            <div className="ipv-more-menu">
                                <button type="button" onClick={() => { handleCalcularConsumo(); setShowMore(false); }}>Calcular Consumo</button>
                                <button type="button" onClick={() => { handleCalcularDiferencias(); setShowMore(false); }}>Calcular Diferencias</button>
                                <button type="button" onClick={() => { onLimpiar(); setShowMore(false); }}>Limpiar Datos</button>
                            </div>
                        )}
                    </div>
                    <button type="button" className="btn btn-outline-secondary d-inline-flex align-items-center gap-2" onClick={onGenerarReporte} disabled={loading}>
                        <DownloadIcon size={16} /> Exportar PDF
                    </button>
                    <button type="button" className="btn btn-primary d-inline-flex align-items-center gap-2" onClick={onGuardar} disabled={loading}>
                        {loading ? <Spinner as="span" animation="border" size="sm" /> : <CheckIcon size={16} />}
                        Guardar cambios
                    </button>
                </div>
            </div>

            {areas.length > 0 && (
                <div className="d-flex align-items-center flex-wrap gap-2 py-2 px-4 border-bottom"
                    style={{ position: 'sticky', top: '73px', zIndex: 9, backgroundColor: 'var(--color-base)' }}>
                    {areas.map(areaNombre => {
                        const Icon = iconoDeArea(areaNombre);
                        const isActive = areaNombre === activeArea;
                        return (
                            <button
                                key={areaNombre}
                                type="button"
                                className={`ipv-tab${isActive ? ' active' : ''}`}
                                onClick={() => seleccionarArea(areaNombre)}
                            >
                                <Icon size={15} />
                                {areaNombre}
                            </button>
                        );
                    })}
                    <div className="ms-auto d-flex align-items-center gap-2" style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>
                        <LockIcon size={13} style={{ opacity: 0.7 }} /> = calculado automáticamente
                    </div>
                </div>
            )}

            <div className="p-4">
                {error && <Alert variant="danger">{error}</Alert>}

                {loading && areas.length === 0 && (
                    <div className="text-center py-5">
                        <Spinner animation="border" role="status">
                            <span className="visually-hidden">Cargando...</span>
                        </Spinner>
                    </div>
                )}

                {activeArea && inventario[activeArea] && (
                    <AreaIPVTab
                        areaData={inventario[activeArea]}
                        onItemChange={(productoId, field, value) => onItemChange(activeArea, productoId, field, value)}
                        onCommentChange={(productoId, field, comment) => onCommentChange(activeArea, productoId, field, comment)}
                    />
                )}

                {!loading && areas.length === 0 && !error && (
                    <Alert variant="secondary">
                        No hay áreas con modelo de IPV configurado. Ve a "Crear Modelos de Área" desde Control IPV.
                    </Alert>
                )}
            </div>
        </div>
    );
}

export default RegistroIPV;
