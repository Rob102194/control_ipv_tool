from flask import Blueprint, request, jsonify
from dependency_injector.wiring import inject, Provide
from src.infrastructure.container import Container
from src.application.use_cases.area_use_cases import (
    CrearAreaUseCase,
    ObtenerAreasUseCase,
    ObtenerAreaPorIdUseCase,
    ActualizarAreaUseCase,
    EliminarAreaUseCase
)

# Creación del Blueprint para las rutas de áreas
area_bp = Blueprint('area', __name__, url_prefix='/api/areas/')

# Ruta para crear una nueva área
@area_bp.route('/', methods=['POST'])
@inject
def crear_area(crear_uc: CrearAreaUseCase = Provide[Container.crear_area_uc]):
    """
    Crea una nueva área a partir de los datos JSON de la solicitud.
    Utiliza el caso de uso de creación de áreas inyectado.
    """
    data = request.get_json()
    try:
        area = crear_uc.execute(data)
        return jsonify(area.to_dict()), 201
    except Exception as e:
        return jsonify({"error": str(e)}), 400

# Ruta para obtener todas las áreas
@area_bp.route('/', methods=['GET'])
@inject
def obtener_areas(obtener_uc: ObtenerAreasUseCase = Provide[Container.obtener_areas_uc]):
    """
    Obtiene una lista de todas las áreas.
    Utiliza el caso de uso de obtención de áreas inyectado.
    """
    try:
        areas = obtener_uc.execute()
        return jsonify([a.to_dict() for a in areas]), 200
    except Exception as e:
        return jsonify({"error": str(e)}), 400

# Ruta para obtener un área por su ID
@area_bp.route('/<id>/', methods=['GET'])
@inject
def obtener_area(id, obtener_por_id_uc: ObtenerAreaPorIdUseCase = Provide[Container.obtener_area_por_id_uc]):
    """
    Obtiene un área específica por su ID.
    Utiliza el caso de uso de obtención por ID inyectado.
    """
    try:
        area = obtener_por_id_uc.execute(id)
        if not area:
            return jsonify({"error": "Área no encontrada"}), 404
        return jsonify(area.to_dict()), 200
    except Exception as e:
        return jsonify({"error": str(e)}), 400

# Ruta para actualizar un área existente
@area_bp.route('/<id>/', methods=['PUT'])
@inject
def actualizar_area(id, actualizar_uc: ActualizarAreaUseCase = Provide[Container.actualizar_area_uc]):
    """
    Actualiza un área existente por su ID con los datos JSON proporcionados.
    Utiliza el caso de uso de actualización de áreas inyectado.
    """
    data = request.get_json()
    try:
        area = actualizar_uc.execute(id, data)
        if not area:
            return jsonify({"error": "Área no encontrada"}), 404
        return jsonify(area.to_dict()), 200
    except Exception as e:
        return jsonify({"error": str(e)}), 400

# Ruta para eliminar un área
@area_bp.route('/<id>/', methods=['DELETE'])
@inject
def eliminar_area(id, eliminar_uc: EliminarAreaUseCase = Provide[Container.eliminar_area_uc]):
    """
    Elimina un área por su ID.
    Utiliza el caso de uso de eliminación de áreas inyectado.
    """
    try:
        success = eliminar_uc.execute(id)
        if not success:
            return jsonify({"error": "Área no encontrada"}), 404
        return '', 204
    except Exception as e:
        return jsonify({"error": str(e)}), 400
