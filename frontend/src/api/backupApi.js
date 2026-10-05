import apiClient from './client';

// API para copias de seguridad de la base de datos.
const backupApi = {
  listar: () => apiClient.get('/backups'),
  crear: () => apiClient.post('/backups'),
  descargar: (nombre) => apiClient.get(`/backups/${encodeURIComponent(nombre)}/download`, {
    responseType: 'blob',
  }),
  restaurar: (nombre) => apiClient.post(`/backups/${encodeURIComponent(nombre)}/restore`),
};

export default backupApi;
