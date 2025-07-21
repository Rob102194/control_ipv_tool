from flask import Blueprint, request, jsonify
from dependency_injector.wiring import inject, Provide
from src.infrastructure.container import Container
import pandas as pd
from flask import send_file
import io
import re
from openpyxl import Workbook
from openpyxl.worksheet.datavalidation import DataValidation
from src.application.use_cases.receta_use_cases import (
    CrearRecetaUseCase,
    ObtenerRecetasUseCase,
    ObtenerRecetaPorIdUseCase,
    ActualizarRecetaUseCase,
    EliminarRecetaUseCase,
    ImportarRecetasUseCase,
    ExportRecetasExcel,
    ImportRecetasExcel
)
from src.application.use_cases.producto_use_cases import ObtenerProductosUseCase
from src.application.use_cases.area_use_cases import ObtenerAreasUseCase

# Creación del Blueprint para las rutas de recetas
receta_bp = Blueprint('receta', __name__, url_prefix='/api/recetas')

# Ruta para crear una nueva receta
@receta_bp.route('/', methods=['POST'])
@inject
def crear_receta(crear_uc: CrearRecetaUseCase = Provide[Container.crear_receta_uc]):
    """
    Crea una nueva receta a partir de los datos proporcionados en la solicitud.
    Utiliza el caso de uso CrearRecetaUseCase para la lógica de negocio.
    """
    data = request.get_json()
    try:
        receta = crear_uc.execute(data)
        return jsonify(receta.to_dict()), 201
    except Exception as e:
        return jsonify({"error": str(e)}), 400

# Ruta para obtener todas las recetas
@receta_bp.route('/', methods=['GET'])
@inject
def obtener_recetas(obtener_uc: ObtenerRecetasUseCase = Provide[Container.obtener_recetas_uc]):
    """
    Obtiene una lista de todas las recetas existentes.
    Utiliza el caso de uso ObtenerRecetasUseCase.
    """
    try:
        sort_by = request.args.get('sort_by', 'nombre')
        filter_by = request.args.get('filter_by', None)
        recetas = obtener_uc.execute(sort_by=sort_by, filter_by=filter_by)
        return jsonify([r.to_dict() for r in recetas]), 200
    except Exception as e:
        return jsonify({"error": str(e)}), 400

# Ruta para obtener una receta por su ID
@receta_bp.route('/<id>/', methods=['GET'])
@inject
def obtener_receta(id, obtener_por_id_uc: ObtenerRecetaPorIdUseCase = Provide[Container.obtener_receta_por_id_uc]):
    """
    Obtiene los detalles de una receta específica por su ID.
    Utiliza el caso de uso ObtenerRecetaPorIdUseCase.
    """
    try:
        receta = obtener_por_id_uc.execute(id)
        if not receta:
            return jsonify({"error": "Receta no encontrada"}), 404
        return jsonify(receta.to_dict()), 200
    except Exception as e:
        return jsonify({"error": str(e)}), 400

# Ruta para actualizar una receta existente
@receta_bp.route('/<id>/', methods=['PUT'])
@inject
def actualizar_receta(id, actualizar_uc: ActualizarRecetaUseCase = Provide[Container.actualizar_receta_uc]):
    """
    Actualiza los datos de una receta existente por su ID.
    Utiliza el caso de uso ActualizarRecetaUseCase.
    """
    data = request.get_json()
    try:
        receta = actualizar_uc.execute(id, data)
        if not receta:
            return jsonify({"error": "Receta no encontrada"}), 404
        return jsonify(receta.to_dict()), 200
    except Exception as e:
        return jsonify({"error": str(e)}), 400

# Ruta para eliminar una receta
@receta_bp.route('/<id>/', methods=['DELETE'])
@inject
def eliminar_receta(id, eliminar_uc: EliminarRecetaUseCase = Provide[Container.eliminar_receta_uc]):
    """
    Elimina una receta por su ID.
    Utiliza el caso de uso EliminarRecetaUseCase.
    """
    try:
        success = eliminar_uc.execute(id)
        if not success:
            return jsonify({"error": "Receta no encontrada"}), 404
        return '', 204
    except Exception as e:
        return jsonify({"error": str(e)}), 400

# Ruta para importar recetas desde un archivo Excel
@receta_bp.route('/import/', methods=['POST'])
@inject
def import_recetas(
    import_uc=Provide[Container.import_recetas_excel]
):
    """
    Importa recetas desde un archivo Excel.
    """
    if 'file' not in request.files:
        return jsonify({"error": "No se encontró el archivo"}), 400
    
    file = request.files['file']
    if file.filename == '':
        return jsonify({"error": "No se seleccionó ningún archivo"}), 400
    
    try:
        import_uc.execute(file)
        return jsonify({"message": "Recetas importadas correctamente"}), 200
    except Exception as e:
        return jsonify({"error": str(e)}), 400

# Ruta para descargar una plantilla de Excel para importar recetas
@receta_bp.route('/export/', methods=['GET'])
@inject
def export_recetas(
    export_uc=Provide[Container.export_recetas_excel]
):
    """
    Exporta todas las recetas a un archivo Excel.
    """
    try:
        excel_file = export_uc.execute()
        return send_file(
            excel_file,
            as_attachment=True,
            download_name='recetas.xlsx',
            mimetype='application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
        )
    except Exception as e:
        return jsonify({"error": str(e)}), 500
