from flask import Blueprint, request, jsonify
from dependency_injector.wiring import inject, Provide
from datetime import datetime
from src.infrastructure.container import Container
from src.application.use_cases.inventario_diario_use_cases import (
    ObtenerEstadoInventarioDiarioUseCase,
    CalcularConsumoUseCase,
    GuardarInventarioDiarioUseCase,
    ObtenerModelosIPVUseCase,
    GuardarModeloIPVUseCase,
    ObtenerRegistrosIPVUseCase,
    GenerarReporteIPVUseCase
)

# Blueprint para los endpoints del inventario diario.
inventario_diario_bp = Blueprint('inventario_diario_bp', __name__, url_prefix='/api/ipv')

# Endpoint para obtener el estado del inventario para una fecha.
@inventario_diario_bp.route('/estado', methods=['GET'])
@inject
def obtener_estado(
    use_case: ObtenerEstadoInventarioDiarioUseCase = Provide[Container.obtener_estado_inventario_diario_uc]
):
    fecha_str = request.args.get('fecha')
    if not fecha_str:
        return jsonify({"error": "El parámetro 'fecha' es requerido"}), 400
    
    try:
        fecha = datetime.strptime(fecha_str, '%Y-%m-%d').date()
        estado = use_case.execute(fecha)
        # Convierte los objetos de dominio a diccionarios.
        estado_dict = {area_nombre: [item.to_dict() for item in items] for area_nombre, items in estado.items()}
        return jsonify(estado_dict), 200
    except ValueError:
        return jsonify({"error": "Formato de fecha inválido. Use YYYY-MM-DD"}), 400
    except Exception as e:
        return jsonify({"error": f"Error inesperado: {e}"}), 500

# Endpoint para calcular el consumo basado en las ventas del día.
@inventario_diario_bp.route('/calcular-consumo', methods=['GET'])
@inject
def calcular_consumo(
    use_case: CalcularConsumoUseCase = Provide[Container.calcular_consumo_uc]
):
    fecha_str = request.args.get('fecha')
    if not fecha_str:
        return jsonify({"error": "El parámetro 'fecha' es requerido"}), 400

    try:
        fecha = datetime.strptime(fecha_str, '%Y-%m-%d').date()
        consumo = use_case.execute(fecha)
        return jsonify(consumo), 200
    except ValueError:
        return jsonify({"error": "Formato de fecha inválido. Use YYYY-MM-DD"}), 400
    except Exception as e:
        return jsonify({"error": f"Error inesperado: {e}"}), 500

# Endpoint para guardar el registro completo del inventario diario.
@inventario_diario_bp.route('/guardar', methods=['POST'])
@inject
def guardar_inventario(
    use_case: GuardarInventarioDiarioUseCase = Provide[Container.guardar_inventario_diario_uc]
):
    data = request.get_json()
    if not data:
        return jsonify({"error": "Cuerpo de la solicitud vacío"}), 400

    try:
        registros_guardados = use_case.execute(data)
        return jsonify([r.to_dict() for r in registros_guardados]), 201
    except Exception as e:
        return jsonify({"error": f"Error al guardar el inventario: {e}"}), 500

# Endpoint para obtener los modelos de IPV.
@inventario_diario_bp.route('/modelos', methods=['GET'])
@inject
def obtener_modelos(
    use_case: ObtenerModelosIPVUseCase = Provide[Container.obtener_modelos_ipv_uc]
):
    try:
        modelos = use_case.execute()
        return jsonify(modelos), 200
    except Exception as e:
        return jsonify({"error": f"Error al obtener los modelos: {e}"}), 500

# Endpoint para guardar un modelo de IPV.
@inventario_diario_bp.route('/modelos', methods=['POST'])
@inject
def guardar_modelo(
    use_case: GuardarModeloIPVUseCase = Provide[Container.guardar_modelo_ipv_uc]
):
    data = request.get_json()
    if not data or 'area_id' not in data or 'productos' not in data:
        return jsonify({"error": "Datos incompletos para guardar el modelo"}), 400

    try:
        modelo = use_case.execute(data['area_id'], data['productos'])
        return jsonify(modelo), 201
    except Exception as e:
        return jsonify({"error": f"Error al guardar el modelo: {e}"}), 500

# Endpoint para obtener todos los registros de IPV.
@inventario_diario_bp.route('/registros', methods=['GET'])
@inject
def obtener_registros(
    use_case: ObtenerRegistrosIPVUseCase = Provide[Container.obtener_registros_ipv_uc]
):
    try:
        registros = use_case.execute()
        return jsonify(registros), 200
    except Exception as e:
        return jsonify({"error": f"Error al obtener los registros: {e}"}), 500

# Endpoint para generar un reporte de IPV.
@inventario_diario_bp.route('/reporte', methods=['GET'])
@inject
def generar_reporte(
    use_case: GenerarReporteIPVUseCase = Provide[Container.generar_reporte_ipv_uc]
):
    fecha_str = request.args.get('fecha')
    if not fecha_str:
        return jsonify({"error": "El parámetro 'fecha' es requerido"}), 400
    
    try:
        fecha = datetime.strptime(fecha_str, '%Y-%m-%d').date()
        reporte_data = use_case.execute(fecha)
        return jsonify(reporte_data), 200
    except ValueError as ve:
        return jsonify({"error": str(ve)}), 404
    except Exception as e:
        return jsonify({"error": f"Error inesperado al generar el reporte: {e}"}), 500
