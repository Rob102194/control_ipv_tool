import apiClient from './client';

// Define un objeto que encapsula todas las llamadas a la API relacionadas con las áreas
const areaApi = {
  /**
   * Envía una solicitud para crear una nueva área.
   * @param {object} area - El objeto del área a crear.
   * @returns {Promise} - La promesa de la respuesta de la API.
   */
  crear: (area) => apiClient.post('areas/', area),

  /**
   * Obtiene todas las áreas.
   * @returns {Promise} - La promesa con la lista de áreas.
   */
  obtenerTodos: () => apiClient.get('areas/'),

  /**
   * Obtiene un área específica por su ID.
   * @param {string} id - El ID del área.
   * @returns {Promise} - La promesa con los datos del área.
   */
  obtenerPorId: (id) => apiClient.get(`areas/${id}/`),

  /**
   * Actualiza un área existente.
   * @param {string} id - El ID del área a actualizar.
   * @param {object} area - Los nuevos datos del área.
   * @returns {Promise} - La promesa de la respuesta de la API.
   */
  actualizar: (id, area) => apiClient.put(`areas/${id}/`, area),

  /**
   * Elimina un área por su ID.
   * @param {string} id - El ID del área a eliminar.
   * @returns {Promise} - La promesa de la respuesta de la API.
   */
  eliminar: (id) => apiClient.delete(`areas/${id}/`),
};

// Exporta el objeto para su uso en otros componentes
export default areaApi;
