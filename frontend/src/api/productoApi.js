import apiClient from './client';
import { fileToBase64 } from './base64';

/**
 * Envía una solicitud para crear un nuevo producto.
 * @param {object} producto - El objeto del producto a crear.
 * @returns {Promise} - La promesa de la respuesta de la API.
 */
export const crearProducto = (producto) => apiClient.post('productos/', producto);

/**
 * Obtiene todos los productos.
 * @returns {Promise} - La promesa con la lista de productos.
 */
export const obtenerProductos = (params) => apiClient.get('productos/', { params });

/**
 * Obtiene un producto específico por su ID.
 * @param {string} id - El ID del producto.
 * @returns {Promise} - La promesa con los datos del producto.
 */
export const obtenerProductoPorId = (id) => apiClient.get(`productos/${id}/`);

/**
 * Actualiza un producto existente.
 * @param {string} id - El ID del producto a actualizar.
 * @param {object} producto - Los nuevos datos del producto.
 * @returns {Promise} - La promesa de la respuesta de la API.
 */
export const actualizarProducto = (id, producto) => apiClient.put(`productos/${id}/`, producto);

/**
 * Elimina un producto por su ID.
 * @param {string} id - El ID del producto a eliminar.
 * @returns {Promise} - La promesa de la respuesta de la API.
 */
export const eliminarProducto = (id) => apiClient.delete(`productos/${id}/`);

/**
 * Exporta los productos a un archivo Excel.
 * @returns {Promise} - La promesa con el archivo Excel.
 */
export const exportarProductos = () => apiClient.get('productos/export/', {
  responseType: 'blob', // Indica que la respuesta es un archivo binario
});

/**
 * Importa productos desde un archivo Excel.
 * @param {File} file - El archivo (.xlsx) a importar.
 * @returns {Promise} - La promesa de la respuesta de la API.
 */
export const importarProductos = async (file) => {
  const archivo_base64 = await fileToBase64(file);
  return apiClient.post('productos/import/', { archivo_base64, nombre_archivo: file.name });
};
