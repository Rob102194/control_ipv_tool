// Importa el cliente de API configurado (axios)
import apiClient from './client';

// Define un objeto que encapsula todas las llamadas a la API relacionadas con las recetas
const recetaApi = {
  /**
   * Envía una solicitud para crear una nueva receta.
   * @param {object} receta - El objeto de la receta a crear.
   * @returns {Promise} - La promesa de la respuesta de la API.
   */
  crear: (receta) => apiClient.post('recetas/', receta),

  /**
   * Obtiene todas las recetas.
   * @returns {Promise} - La promesa con la lista de recetas.
   */
  obtenerTodos: (params) => apiClient.get('recetas/', { params }),

  /**
   * Obtiene una receta específica por su ID.
   * @param {string} id - El ID de la receta.
   * @returns {Promise} - La promesa con los datos de la receta.
   */
  obtenerPorId: (id) => apiClient.get(`recetas/${id}/`),

  /**
   * Actualiza una receta existente.
   * @param {string} id - El ID de la receta a actualizar.
   * @param {object} receta - Los nuevos datos de la receta.
   * @returns {Promise} - La promesa de la respuesta de la API.
   */
  actualizar: (id, receta) => apiClient.put(`recetas/${id}/`, receta),

  /**
   * Elimina una receta por su ID.
   * @param {string} id - El ID de la receta a eliminar.
   * @returns {Promise} - La promesa de la respuesta de la API.
   */
  eliminar: (id) => apiClient.delete(`recetas/${id}/`),

  /**
   * Importa recetas desde un archivo.
   * @param {File} file - El archivo (.xlsx) a importar.
   * @returns {Promise} - La promesa de la respuesta de la API.
   */
  importar: (file) => {
    const formData = new FormData();
    formData.append('file', file);
    return apiClient.post('recetas/import/', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    });
  },

  /**
   * Exporta las recetas a un archivo Excel.
   * @returns {Promise} - La promesa con el archivo Excel.
   */
  exportar: () => apiClient.get('recetas/export/', {
    responseType: 'blob', // Indica que la respuesta es un archivo binario
  }),
};

// Exporta el objeto para su uso en otros componentes
export default recetaApi;
