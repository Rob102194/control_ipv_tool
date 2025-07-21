from flask import Blueprint, request, jsonify, current_app, send_file
from dependency_injector.wiring import inject, Provide
from src.infrastructure.container import Container
from src.core.domain.producto import Producto

# Creación del Blueprint para las rutas de productos
producto_bp = Blueprint('producto', __name__, url_prefix='/api/productos/')

# Ruta para crear un nuevo producto
@producto_bp.route('/', methods=['POST'])
@inject
def crear_producto(
    crear_uc = Provide[Container.crear_producto_uc]
):
    """
    Crea un nuevo producto a partir de los datos JSON de la solicitud.
    Utiliza el caso de uso de creación de productos inyectado.
    """
    data = request.get_json()
    try:
        producto = crear_uc.execute(data)
        return jsonify(producto.to_dict()), 201
    except Exception as e:
        return jsonify({"error": str(e)}), 400

# Ruta para obtener todos los productos
@producto_bp.route('/', methods=['GET'])
@inject
def obtener_productos(
    obtener_uc = Provide[Container.obtener_productos_uc]
):
    """
    Obtiene una lista de todos los productos.
    Utiliza el caso de uso de obtención de productos inyectado.
    """
    try:
        sort_by = request.args.get('sort_by', 'nombre')
        productos = obtener_uc.execute(sort_by=sort_by)
        productos_dict = [p.to_dict() for p in productos]
        return jsonify(productos_dict), 200
    except Exception as e:
        return jsonify({"error": str(e)}), 400

# Ruta para obtener un producto por su ID
@producto_bp.route('/<id>/', methods=['GET'])
@inject
def obtener_producto(
    id,
    obtener_por_id_uc = Provide[Container.obtener_producto_por_id_uc]
):
    """
    Obtiene un producto específico por su ID.
    Utiliza el caso de uso de obtención por ID inyectado.
    """
    try:
        producto = obtener_por_id_uc.execute(id)
        if not producto:
            return jsonify({"error": "Producto no encontrado"}), 404
        return jsonify(producto.to_dict()), 200
    except Exception as e:
        return jsonify({"error": str(e)}), 400

# Ruta para actualizar un producto existente
@producto_bp.route('/<id>/', methods=['PUT'])
@inject
def actualizar_producto(
    id,
    actualizar_uc = Provide[Container.actualizar_producto_uc]
):
    """
    Actualiza un producto existente por su ID con los datos JSON proporcionados.
    Utiliza el caso de uso de actualización de productos inyectado.
    """
    data = request.get_json()
    try:
        producto = actualizar_uc.execute(id, data)
        if not producto:
            return jsonify({"error": "Producto no encontrado"}), 404
        return jsonify(producto.to_dict()), 200
    except Exception as e:
        return jsonify({"error": str(e)}), 400

# Ruta para eliminar un producto
@producto_bp.route('/<id>/', methods=['DELETE'])
@inject
def eliminar_producto(
    id,
    eliminar_uc = Provide[Container.eliminar_producto_uc]
):
    """
    Elimina un producto por su ID.
    Utiliza el caso de uso de eliminación de productos inyectado.
    """
    try:
        success = eliminar_uc.execute(id)
        if not success:
            return jsonify({"error": "Producto no encontrado"}), 404
        return '', 204
    except Exception as e:
        return jsonify({"error": str(e)}), 400

# Ruta para exportar productos a Excel
@producto_bp.route('/export/', methods=['GET'])
@inject
def export_productos(
    export_uc=Provide[Container.export_productos_excel]
):
    """
    Exporta todos los productos a un archivo Excel.
    """
    try:
        excel_file = export_uc.execute()
        return send_file(
            excel_file,
            as_attachment=True,
            download_name='productos.xlsx',
            mimetype='application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
        )
    except Exception as e:
        return jsonify({"error": str(e)}), 500

# Ruta para importar productos desde Excel
@producto_bp.route('/import/', methods=['POST'])
@inject
def import_productos(
    import_uc=Provide[Container.import_productos_excel]
):
    """
    Importa productos desde un archivo Excel.
    """
    if 'file' not in request.files:
        return jsonify({"error": "No se encontró el archivo"}), 400
    
    file = request.files['file']
    if file.filename == '':
        return jsonify({"error": "No se seleccionó ningún archivo"}), 400
    
    try:
        import_uc.execute(file)
        return jsonify({"message": "Productos importados correctamente"}), 200
    except Exception as e:
        return jsonify({"error": str(e)}), 400
