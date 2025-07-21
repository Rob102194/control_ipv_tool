from flask import Blueprint, jsonify
from dependency_injector.wiring import inject, Provide
from src.infrastructure.container import Container
from src.application.use_cases.historial_use_cases import ObtenerHistorialUseCase

historial_bp = Blueprint('historial', __name__, url_prefix='/api/historial')

@historial_bp.route('/<entidad_tipo>/', methods=['GET'])
@inject
def obtener_historial(
    entidad_tipo: str,
    obtener_historial_uc: ObtenerHistorialUseCase = Provide[Container.obtener_historial_uc]
):
    try:
        historial = obtener_historial_uc.execute(entidad_tipo)
        return jsonify([h.to_dict() for h in historial]), 200
    except Exception as e:
        return jsonify({"error": str(e)}), 400
