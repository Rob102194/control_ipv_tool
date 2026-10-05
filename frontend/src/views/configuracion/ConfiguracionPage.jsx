import React, { useEffect, useState } from 'react';
import { Container, Button, Form, Spinner, Alert } from 'react-bootstrap';
import configuracionApi from '../../api/configuracionApi';
import backupApi from '../../api/backupApi';
import { useToast } from '../../contexts/ToastContext';
import { useConfirm } from '../../contexts/ConfirmContext';
import { useRestoreLock } from '../../contexts/RestoreLockContext';
import { formatDateTimeEs } from '../../utils/date';
import { CheckIcon, ArchiveIcon, DownloadIcon, RefreshIcon, PlusIcon, FolderIcon, ChevronDownIcon } from '../../components/icons';

function formatearTamano(bytes) {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

// Página de ajustes del negocio: pensada como el único lugar donde viven
// (y vivirán) todas las opciones de configuración de la app, accesible desde
// el ícono de tuerca en la barra de navegación.
const ConfiguracionPage = () => {
  const showToast = useToast();
  const confirmar = useConfirm();
  const bloquearPorRestauracion = useRestoreLock();

  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const [nombre, setNombre] = useState('');
  const [guardandoNombre, setGuardandoNombre] = useState(false);

  const [backupDir, setBackupDir] = useState('');
  const [guardandoBackupDir, setGuardandoBackupDir] = useState(false);
  const [eligiendoCarpeta, setEligiendoCarpeta] = useState(false);

  const [backups, setBackups] = useState([]);
  const [cargandoBackups, setCargandoBackups] = useState(true);
  const [creandoBackup, setCreandoBackup] = useState(false);
  const [restaurando, setRestaurando] = useState(null); // nombre del backup en curso, o null
  const [mostrarBackups, setMostrarBackups] = useState(false);

  const cargarConfiguracion = async () => {
    try {
      setLoading(true);
      const { data } = await configuracionApi.obtener();
      setNombre(data.nombre_negocio || '');
      setBackupDir(data.backup_dir || '');
      setError('');
    } catch (err) {
      setError('Error al cargar la configuración.');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const cargarBackups = async () => {
    try {
      setCargandoBackups(true);
      const { data } = await backupApi.listar();
      setBackups(data);
    } catch (err) {
      showToast('Error al cargar la lista de backups.', 'danger');
      console.error(err);
    } finally {
      setCargandoBackups(false);
    }
  };

  useEffect(() => {
    cargarConfiguracion();
    cargarBackups();
  }, []);

  const guardarNombre = async () => {
    setGuardandoNombre(true);
    try {
      const { data } = await configuracionApi.actualizarNombre(nombre);
      // La etiqueta de la barra de navegación (NegocioBadge, en App.jsx) solo
      // carga el nombre una vez al montar y no se remonta al navegar dentro
      // de la SPA: sin avisarle, seguiría mostrando el nombre viejo hasta
      // recargar la página entera.
      window.dispatchEvent(new CustomEvent('negocio-actualizado', { detail: { nombre: data.nombre_negocio } }));
      showToast('Nombre del negocio actualizado.');
    } catch (err) {
      showToast(err.response?.data?.error || 'Error al guardar el nombre.', 'danger');
    } finally {
      setGuardandoNombre(false);
    }
  };

  const elegirCarpeta = async () => {
    setEligiendoCarpeta(true);
    try {
      const { data } = await backupApi.elegirCarpeta();
      // El usuario puede cancelar el diálogo nativo: el backend devuelve
      // path vacío (no es un error), así que simplemente no tocamos nada.
      if (!data.path) return;
      setBackupDir(data.path);
      // A diferencia de escribir la ruta a mano, elegirla con el diálogo es
      // una selección deliberada y completa (nunca un valor a medio
      // escribir) — se guarda de inmediato para que "Backup ahora" la use
      // ya mismo, sin depender de que el usuario pulse "Guardar" aparte.
      await configuracionApi.actualizarBackupDir(data.path);
      showToast('Carpeta de backups actualizada.');
    } catch (err) {
      if (err.response?.status === 404) {
        showToast('El selector de carpetas no está disponible en este modo; escribe la ruta a mano.', 'danger');
      } else {
        showToast(err.response?.data?.error || 'Error al abrir el selector de carpetas.', 'danger');
      }
    } finally {
      setEligiendoCarpeta(false);
    }
  };

  const guardarBackupDir = async () => {
    setGuardandoBackupDir(true);
    try {
      await configuracionApi.actualizarBackupDir(backupDir);
      showToast('Carpeta de backups actualizada.');
    } catch (err) {
      showToast(err.response?.data?.error || 'Error al guardar la carpeta de backups.', 'danger');
    } finally {
      setGuardandoBackupDir(false);
    }
  };

  const crearBackup = async () => {
    setCreandoBackup(true);
    try {
      await backupApi.crear();
      showToast('Backup creado correctamente.');
      await cargarBackups();
    } catch (err) {
      showToast(err.response?.data?.error || 'Error al crear el backup.', 'danger');
    } finally {
      setCreandoBackup(false);
    }
  };

  const descargarBackup = async (nombreArchivo) => {
    try {
      const response = await backupApi.descargar(nombreArchivo);
      const url = window.URL.createObjectURL(new Blob([response.data]));
      const link = document.createElement('a');
      link.href = url;
      link.setAttribute('download', nombreArchivo);
      document.body.appendChild(link);
      link.click();
      link.remove();
    } catch (err) {
      showToast('Error al descargar el backup.', 'danger');
      console.error(err);
    }
  };

  const restaurarBackup = async (nombreArchivo) => {
    const ok = await confirmar(
      `Esto reemplazará TODOS los datos actuales por los del backup del ${formatDateTimeEs(
        backups.find((b) => b.nombre === nombreArchivo)?.creado_en
      )}. Se guardará un backup del estado actual antes de continuar, pero después deberás cerrar y volver a abrir la aplicación. ¿Continuar?`,
      { confirmLabel: 'Restaurar' }
    );
    if (!ok) return;
    setRestaurando(nombreArchivo);
    try {
      await backupApi.restaurar(nombreArchivo);
      bloquearPorRestauracion();
    } catch (err) {
      showToast(err.response?.data?.error || 'Error al restaurar el backup.', 'danger');
      setRestaurando(null);
    }
  };

  if (loading) {
    return (
      <Container className="mt-5 text-center">
        <Spinner animation="border" role="status">
          <span className="visually-hidden">Cargando...</span>
        </Spinner>
      </Container>
    );
  }

  return (
    <Container className="mt-4">
      <div className="page-header">
        <div>
          <h1>Configuración</h1>
          <p>Ajustes del negocio y copias de seguridad de la base de datos.</p>
        </div>
        <div className="page-actions">
          <Button
            variant="primary"
            onClick={crearBackup}
            disabled={creandoBackup}
            className="d-inline-flex align-items-center gap-2"
          >
            {creandoBackup ? <Spinner animation="border" size="sm" /> : <PlusIcon size={15} />}
            Backup ahora
          </Button>
        </div>
      </div>

      {error && <Alert variant="danger" onClose={() => setError('')} dismissible>{error}</Alert>}

      <div className="card mb-4" style={{ maxWidth: '560px' }}>
        <h2 style={{ fontSize: '0.9375rem', fontWeight: 700, marginBottom: '14px' }}>Negocio</h2>
        <Form.Group className="mb-3">
          <Form.Label className="fw-semibold">Nombre</Form.Label>
          <Form.Control
            type="text"
            value={nombre}
            onChange={(e) => setNombre(e.target.value)}
            placeholder="Ej: Pizzería Don Mario"
            maxLength={80}
          />
          <Form.Text style={{ color: 'var(--text-secondary)' }}>
            Útil si operas más de un negocio con instalaciones separadas de la app: ayuda a distinguir cada
            ventana de un vistazo.
          </Form.Text>
        </Form.Group>
        <Button
          variant="primary"
          onClick={guardarNombre}
          disabled={guardandoNombre}
          className="d-inline-flex align-items-center gap-2"
        >
          {guardandoNombre ? <Spinner animation="border" size="sm" /> : <CheckIcon size={15} />}
          Guardar
        </Button>
      </div>

      <div className="card mb-4" style={{ maxWidth: '560px' }}>
        <h2 style={{ fontSize: '0.9375rem', fontWeight: 700, marginBottom: '14px' }} className="d-flex align-items-center gap-2">
          <ArchiveIcon size={16} /> Carpeta de backups
        </h2>
        <Form.Group className="mb-3">
          <Form.Label className="fw-semibold">Ruta</Form.Label>
          <div className="d-flex gap-2">
            <Form.Control
              type="text"
              value={backupDir}
              onChange={(e) => setBackupDir(e.target.value)}
              placeholder="Dejar en blanco para usar la carpeta por defecto"
            />
            <Button
              variant="outline-secondary"
              onClick={elegirCarpeta}
              disabled={eligiendoCarpeta}
              className="d-inline-flex align-items-center gap-2 flex-shrink-0"
            >
              {eligiendoCarpeta ? <Spinner animation="border" size="sm" /> : <FolderIcon size={15} />}
              Elegir carpeta…
            </Button>
          </div>
          <Form.Text style={{ color: 'var(--text-secondary)' }}>
            Por defecto se guardan junto a la base de datos. Puedes apuntarla a una carpeta ya sincronizada por
            Dropbox, Google Drive o similar, para tener una copia fuera de esta máquina.
          </Form.Text>
        </Form.Group>
        <Button
          variant="outline-secondary"
          onClick={guardarBackupDir}
          disabled={guardandoBackupDir}
          className="d-inline-flex align-items-center gap-2"
        >
          {guardandoBackupDir ? <Spinner animation="border" size="sm" /> : <CheckIcon size={15} />}
          Guardar
        </Button>
      </div>

      <button
        type="button"
        onClick={() => setMostrarBackups((v) => !v)}
        className="d-flex align-items-center gap-2 w-100 bg-transparent border-0 p-0 mb-2"
        style={{ cursor: 'pointer' }}
        aria-expanded={mostrarBackups}
      >
        <h2 style={{ fontSize: '0.9375rem', fontWeight: 700, margin: 0 }}>
          Backups disponibles{!cargandoBackups && ` (${backups.length})`}
        </h2>
        <span
          className="d-inline-flex"
          style={{
            transition: 'transform 0.15s ease',
            transform: mostrarBackups ? 'rotate(180deg)' : 'none',
            color: 'var(--text-secondary)',
          }}
        >
          <ChevronDownIcon size={16} />
        </span>
      </button>
      {mostrarBackups && (
        cargandoBackups ? (
          <div className="text-center py-4">
            <Spinner animation="border" size="sm" />
          </div>
        ) : backups.length === 0 ? (
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>
            Todavía no hay backups. Usa "Backup ahora" para crear el primero.
          </p>
        ) : (
          <div className="list-card">
            <table className="list-table">
              <thead>
                <tr>
                  <th>Fecha</th>
                  <th>Tamaño</th>
                  <th>Acciones</th>
                </tr>
              </thead>
              <tbody>
                {backups.map((b) => (
                  <tr key={b.nombre}>
                    <td data-label="Fecha" style={{ fontWeight: 600 }}>{formatDateTimeEs(b.creado_en)}</td>
                    <td data-label="Tamaño" style={{ color: 'var(--text-secondary)' }}>{formatearTamano(b.tamano_bytes)}</td>
                    <td data-label="">
                      <Button
                        variant="outline-secondary"
                        size="sm"
                        className="me-2 d-inline-flex align-items-center gap-1"
                        onClick={() => descargarBackup(b.nombre)}
                      >
                        <DownloadIcon size={13} /> Descargar
                      </Button>
                      <Button
                        variant="outline-danger"
                        size="sm"
                        className="d-inline-flex align-items-center gap-1"
                        onClick={() => restaurarBackup(b.nombre)}
                        disabled={restaurando !== null}
                      >
                        {restaurando === b.nombre ? <Spinner animation="border" size="sm" /> : <RefreshIcon size={13} />}
                        Restaurar
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )
      )}
    </Container>
  );
};

export default ConfiguracionPage;
