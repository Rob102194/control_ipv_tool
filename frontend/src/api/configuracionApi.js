import apiClient from './client';

// Ajustes del negocio (una sola fila por base de datos). El backend aplica
// una actualización PARCIAL: un campo ausente del cuerpo JSON no se toca
// (distinto de mandarlo como cadena vacía, que sí lo borra) — por eso cada
// función solo incluye el campo que le corresponde, para no pisar el otro.
const configuracionApi = {
  obtener: () => apiClient.get('/configuracion'),
  actualizarNombre: (nombreNegocio) => apiClient.put('/configuracion', { nombre_negocio: nombreNegocio }),
  actualizarBackupDir: (backupDir) => apiClient.put('/configuracion', { backup_dir: backupDir }),
};

export default configuracionApi;
