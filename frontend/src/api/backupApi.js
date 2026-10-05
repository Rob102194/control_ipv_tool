import apiClient from './client';

// API para copias de seguridad de la base de datos.
const backupApi = {
  listar: () => apiClient.get('/backups'),
  crear: () => apiClient.post('/backups'),
  descargar: (nombre) => apiClient.get(`/backups/${encodeURIComponent(nombre)}/download`, {
    responseType: 'blob',
  }),
  restaurar: (nombre) => apiClient.post(`/backups/${encodeURIComponent(nombre)}/restore`),
  // Selector nativo de carpetas (solo escritorio). En web el backend responde
  // 404 "no disponible en este modo"; el formulario sigue permitiendo
  // escribir la ruta a mano como alternativa universal.
  elegirCarpeta: () => apiClient.post('/backups/elegir-carpeta'),
};

export default backupApi;
