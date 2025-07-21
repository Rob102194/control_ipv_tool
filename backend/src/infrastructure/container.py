from dependency_injector import containers, providers
from src.infrastructure.db.models import db
from src.infrastructure.repositories.sqlite_producto_repository import SQLiteProductoRepository
from src.application.use_cases.producto_use_cases import (
    CrearProductoUseCase,
    ObtenerProductosUseCase,
    ObtenerProductoPorIdUseCase,
    ActualizarProductoUseCase,
    EliminarProductoUseCase,
    ExportProductosExcel,
    ImportProductosExcel
)
from src.infrastructure.repositories.sqlite_area_repository import SQLiteAreaRepository
from src.application.use_cases.area_use_cases import (
    CrearAreaUseCase,
    ObtenerAreasUseCase,
    ObtenerAreaPorIdUseCase,
    ActualizarAreaUseCase,
    EliminarAreaUseCase
)
from src.infrastructure.repositories.sqlite_receta_repository import SQLiteRecetaRepository
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
from src.infrastructure.repositories.sqlite_venta_repository import SQLiteVentaRepository
from src.application.use_cases.venta_use_cases import (
    CrearVentaUseCase,
    ObtenerVentasUseCase,
    ObtenerVentaPorIdUseCase,
    ActualizarVentaUseCase,
    EliminarVentaUseCase,
    ImportarVentasUseCase,
    EliminarVentasMultiplesUseCase
)
from src.infrastructure.repositories.sqlite_inventario_diario_repository import SQLiteInventarioDiarioRepository
from src.application.use_cases.inventario_diario_use_cases import (
    ObtenerEstadoInventarioDiarioUseCase,
    CalcularConsumoUseCase,
    GuardarInventarioDiarioUseCase,
    ObtenerModelosIPVUseCase,
    GuardarModeloIPVUseCase,
    ObtenerRegistrosIPVUseCase,
    GenerarReporteIPVUseCase
)
from src.infrastructure.repositories.sqlite_historial_repository import SQLiteHistorialRepository
from src.application.use_cases.historial_use_cases import (
    RegistrarCambioUseCase,
    ObtenerHistorialUseCase
)

class Container(containers.DeclarativeContainer):
    # Configuración
    config = providers.Configuration()
    
    # Base de datos
    db_session = providers.Singleton(db.session)
    
    # Repositorios
    producto_repository = providers.Factory(
        SQLiteProductoRepository,
        db_session=db_session
    )

    area_repository = providers.Factory(
        SQLiteAreaRepository,
        db_session=db_session
    )

    receta_repository = providers.Factory(
        SQLiteRecetaRepository,
        db_session=db_session
    )

    venta_repository = providers.Factory(
        SQLiteVentaRepository,
        db_session=db_session
    )

    inventario_diario_repository = providers.Factory(
        SQLiteInventarioDiarioRepository,
        db_session=db_session
    )

    historial_repository = providers.Factory(
        SQLiteHistorialRepository,
        db_session=db_session
    )
    
    # Casos de uso para Productos
    crear_producto_uc = providers.Factory(
        CrearProductoUseCase,
        repository=producto_repository,
        registrar_cambio_uc=providers.Factory(RegistrarCambioUseCase, repository=providers.Factory(SQLiteHistorialRepository, db_session=db_session))
    )
    
    obtener_productos_uc = providers.Factory(
        ObtenerProductosUseCase,
        repository=producto_repository
    )
    
    obtener_producto_por_id_uc = providers.Factory(
        ObtenerProductoPorIdUseCase,
        repository=producto_repository
    )
    
    actualizar_producto_uc = providers.Factory(
        ActualizarProductoUseCase,
        repository=producto_repository,
        registrar_cambio_uc=providers.Factory(RegistrarCambioUseCase, repository=providers.Factory(SQLiteHistorialRepository, db_session=db_session))
    )
    
    eliminar_producto_uc = providers.Factory(
        EliminarProductoUseCase,
        repository=producto_repository,
        registrar_cambio_uc=providers.Factory(RegistrarCambioUseCase, repository=providers.Factory(SQLiteHistorialRepository, db_session=db_session))
    )

    export_productos_excel = providers.Factory(
        ExportProductosExcel,
        repository=producto_repository
    )

    import_productos_excel = providers.Factory(
        ImportProductosExcel,
        repository=producto_repository
    )

    # Casos de uso para Áreas
    crear_area_uc = providers.Factory(
        CrearAreaUseCase,
        repository=area_repository
    )
    
    obtener_areas_uc = providers.Factory(
        ObtenerAreasUseCase,
        repository=area_repository
    )
    
    obtener_area_por_id_uc = providers.Factory(
        ObtenerAreaPorIdUseCase,
        repository=area_repository
    )
    
    actualizar_area_uc = providers.Factory(
        ActualizarAreaUseCase,
        repository=area_repository
    )
    
    eliminar_area_uc = providers.Factory(
        EliminarAreaUseCase,
        repository=area_repository
    )

    # Casos de uso para Recetas
    crear_receta_uc = providers.Factory(
        CrearRecetaUseCase,
        repository=receta_repository,
        registrar_cambio_uc=providers.Factory(RegistrarCambioUseCase, repository=providers.Factory(SQLiteHistorialRepository, db_session=db_session))
    )

    obtener_recetas_uc = providers.Factory(
        ObtenerRecetasUseCase,
        repository=receta_repository
    )

    obtener_receta_por_id_uc = providers.Factory(
        ObtenerRecetaPorIdUseCase,
        repository=receta_repository
    )

    actualizar_receta_uc = providers.Factory(
        ActualizarRecetaUseCase,
        repository=receta_repository,
        registrar_cambio_uc=providers.Factory(RegistrarCambioUseCase, repository=providers.Factory(SQLiteHistorialRepository, db_session=db_session))
    )

    eliminar_receta_uc = providers.Factory(
        EliminarRecetaUseCase,
        repository=receta_repository,
        registrar_cambio_uc=providers.Factory(RegistrarCambioUseCase, repository=providers.Factory(SQLiteHistorialRepository, db_session=db_session))
    )

    importar_recetas_uc = providers.Factory(
        ImportarRecetasUseCase,
        repository=receta_repository
    )

    export_recetas_excel = providers.Factory(
        ExportRecetasExcel,
        repository=receta_repository,
        producto_repository=producto_repository,
        area_repository=area_repository
    )

    import_recetas_excel = providers.Factory(
        ImportRecetasExcel,
        repository=receta_repository,
        producto_repository=producto_repository,
        area_repository=area_repository
    )

    # Casos de uso para Ventas
    crear_venta_uc = providers.Factory(
        CrearVentaUseCase,
        repository=venta_repository
    )

    obtener_ventas_uc = providers.Factory(
        ObtenerVentasUseCase,
        repository=venta_repository
    )

    obtener_venta_por_id_uc = providers.Factory(
        ObtenerVentaPorIdUseCase,
        repository=venta_repository
    )

    actualizar_venta_uc = providers.Factory(
        ActualizarVentaUseCase,
        repository=venta_repository
    )

    eliminar_venta_uc = providers.Factory(
        EliminarVentaUseCase,
        repository=venta_repository
    )

    importar_ventas_uc = providers.Factory(
        ImportarVentasUseCase,
        repository=venta_repository,
        receta_repository=receta_repository
    )

    eliminar_ventas_multiples_uc = providers.Factory(
        EliminarVentasMultiplesUseCase,
        repository=venta_repository
    )

    # Casos de uso para Inventario Diario
    obtener_estado_inventario_diario_uc = providers.Factory(
        ObtenerEstadoInventarioDiarioUseCase,
        inventario_repository=inventario_diario_repository,
        producto_repository=producto_repository,
        area_repository=area_repository
    )

    calcular_consumo_uc = providers.Factory(
        CalcularConsumoUseCase,
        venta_repository=venta_repository,
        receta_repository=receta_repository
    )

    guardar_inventario_diario_uc = providers.Factory(
        GuardarInventarioDiarioUseCase,
        inventario_repository=inventario_diario_repository
    )

    obtener_modelos_ipv_uc = providers.Factory(
        ObtenerModelosIPVUseCase,
        inventario_repository=inventario_diario_repository
    )

    guardar_modelo_ipv_uc = providers.Factory(
        GuardarModeloIPVUseCase,
        inventario_repository=inventario_diario_repository
    )

    obtener_registros_ipv_uc = providers.Factory(
        ObtenerRegistrosIPVUseCase,
        inventario_repository=inventario_diario_repository
    )

    generar_reporte_ipv_uc = providers.Factory(
        GenerarReporteIPVUseCase,
        inventario_repository=inventario_diario_repository,
        area_repository=area_repository,
        producto_repository=producto_repository
    )

    # Casos de uso para Historial
    registrar_cambio_uc = providers.Factory(
        RegistrarCambioUseCase,
        repository=historial_repository
    )

    obtener_historial_uc = providers.Factory(
        ObtenerHistorialUseCase,
        repository=historial_repository
    )
