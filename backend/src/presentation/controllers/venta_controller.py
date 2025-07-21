from flask import Blueprint, request, jsonify
from dependency_injector.wiring import inject, Provide
from src.infrastructure.container import Container
from src.core.domain.venta import Venta

# Creación del Blueprint para las rutas de ventas
venta_bp = Blueprint('venta', __name__, url_prefix='/api/ventas/')

# Ruta para crear una nueva venta
@venta_bp.route('/', methods=['POST'])
@inject
def crear_venta(
    crear_uc=Provide[Container.crear_venta_uc]
):
    data = request.get_json()
    try:
        venta = crear_uc.execute(data)
        return jsonify(venta.to_dict()), 201
    except Exception as e:
        return jsonify({"error": str(e)}), 400

# Ruta para obtener todas las ventas
@venta_bp.route('/', methods=['GET'])
@inject
def obtener_ventas(
    obtener_uc=Provide[Container.obtener_ventas_uc]
):
    try:
        ventas = obtener_uc.execute()
        ventas_dict = [v.to_dict() for v in ventas]
        return jsonify(ventas_dict), 200
    except Exception as e:
        return jsonify({"error": str(e)}), 400

# Ruta para obtener una venta por su ID
@venta_bp.route('/<id>/', methods=['GET'])
@inject
def obtener_venta(
    id,
    obtener_por_id_uc=Provide[Container.obtener_venta_por_id_uc]
):
    try:
        venta = obtener_por_id_uc.execute(id)
        if not venta:
            return jsonify({"error": "Venta no encontrada"}), 404
        return jsonify(venta.to_dict()), 200
    except Exception as e:
        return jsonify({"error": str(e)}), 400

# Ruta para actualizar una venta existente
@venta_bp.route('/<id>/', methods=['PUT'])
@inject
def actualizar_venta(
    id,
    actualizar_uc=Provide[Container.actualizar_venta_uc]
):
    data = request.get_json()
    try:
        venta = actualizar_uc.execute(id, data)
        if not venta:
            return jsonify({"error": "Venta no encontrada"}), 404
        return jsonify(venta.to_dict()), 200
    except Exception as e:
        return jsonify({"error": str(e)}), 400

# Ruta para eliminar una venta
@venta_bp.route('/<id>/', methods=['DELETE'])
@inject
def eliminar_venta(
    id,
    eliminar_uc=Provide[Container.eliminar_venta_uc]
):
    try:
        success = eliminar_uc.execute(id)
        if not success:
            return jsonify({"error": "Venta no encontrada"}), 404
        return '', 204
    except Exception as e:
        return jsonify({"error": str(e)}), 400

# Ruta para eliminar múltiples ventas
@venta_bp.route('/delete-multiple/', methods=['POST'])
@inject
def eliminar_ventas_multiples(
    eliminar_multiples_uc=Provide[Container.eliminar_ventas_multiples_uc]
):
    data = request.get_json()
    ids = data.get('ids')
    if not ids:
        return jsonify({"error": "No se proporcionaron IDs para eliminar"}), 400
    try:
        success = eliminar_multiples_uc.execute(ids)
        if not success:
            return jsonify({"error": "No se pudieron eliminar las ventas"}), 404
        return '', 204
    except Exception as e:
        return jsonify({"error": str(e)}), 400

# Ruta para importar ventas desde un archivo
@venta_bp.route('/importar/', methods=['POST'])
@inject
def importar_ventas(
    importar_uc=Provide[Container.importar_ventas_uc]
):
    if 'file' not in request.files:
        return jsonify({"error": "No se encontró el archivo"}), 400
        
    file = request.files['file']
    if file.filename == '':
        return jsonify({"error": "Nombre de archivo inválido"}), 400
        
    fecha = request.form.get('fecha')
    
    try:
        if not file.filename.endswith(('.xlsx', '.xls')):
            return jsonify({"error": "Formato de archivo no soportado"}), 400
            
        ventas, nuevas_recetas = importar_uc.execute(file.stream, fecha)
        return jsonify({
            "message": f"Se importaron {len(ventas)} ventas correctamente",
            "ventas": [v.to_dict() for v in ventas],
            "nuevas_recetas": [r.to_dict() for r in nuevas_recetas]
        }), 200
        
    except ValueError as ve:
        return jsonify({"error": str(ve)}), 400
    except Exception as e:
        return jsonify({"error": f"Error al procesar el archivo: {str(e)}"}), 500
