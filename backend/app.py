import os
from dotenv import load_dotenv
from flask import Flask
from flask_cors import CORS
from flask_migrate import Migrate
from sqlalchemy.exc import OperationalError
from src.infrastructure.db.models import db, Producto, Area, Receta, Ingrediente, MovimientoInventario, Venta, InventarioDiario, ModeloIPV
from src.presentation.controllers import producto_controller
from src.presentation.controllers import area_controller
from src.presentation.controllers import receta_controller
from src.presentation.controllers import venta_controller
from src.presentation.controllers import inventario_diario_controller
from src.presentation.controllers import historial_controller
from src.infrastructure.container import Container

from datetime import datetime

def create_app() -> Flask:
    # Cargar variables de entorno
    dotenv_path = os.path.join(os.path.dirname(__file__), 'env', '.env')
    load_dotenv(dotenv_path=dotenv_path)
    
    print(f"App reloaded at: {datetime.now()}")

    container = Container()
    container.wire(modules=[
        producto_controller, 
        area_controller, 
        receta_controller, 
        venta_controller,
        inventario_diario_controller,
        historial_controller
    ])
    
    app = Flask(__name__)
    app.container = container
    app.config['SQLALCHEMY_DATABASE_URI'] = os.getenv('DB_URI', 'sqlite:///inventario.db')
    app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = False
    
    # Configurar CORS
    CORS(app, resources={r"/api/*": {"origins": "*"}})

    app.register_blueprint(producto_controller.producto_bp)
    app.register_blueprint(area_controller.area_bp)
    app.register_blueprint(receta_controller.receta_bp)
    app.register_blueprint(venta_controller.venta_bp)
    app.register_blueprint(inventario_diario_controller.inventario_diario_bp)
    app.register_blueprint(historial_controller.historial_bp)

    db.init_app(app)
    migrate = Migrate(app, db)
    
    container.db_session.override(db.session)

    with app.app_context():
        try:
            db.create_all()
        except OperationalError as e:
            app.logger.warning(
                "Could not create database tables."
                " This can be ignored if the tables already exist."
                f" Error: {e}"
            )

    @app.route('/')
    def health_check():
        return {'status': 'ok', 'message': 'Inventory API running'}

    return app

app = create_app()

if __name__ == '__main__':
    app.run(debug=True, port=5000)
