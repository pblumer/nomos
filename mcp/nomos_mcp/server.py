"""Nomos MCP Server — Haupt-Einstiegspunkt."""

import asyncio

from mcp.server import Server
from mcp.server.stdio import stdio_server
from mcp.types import InitializationOptions, TextContent, Tool

from .tools.blueprints import (
    BLUEPRINTS_TOOLS,
    get_blueprint_handler,
    list_blueprints_handler,
)
from .tools.cosmos import (
    COSMOS_TOOLS,
    cosmos_info_handler,
    cosmos_init_handler,
)
from .tools.decisions import (
    DECISION_TOOLS,
    create_decision_handler,
    get_decision_dmn_handler,
    get_decision_handler,
    list_decisions_handler,
    update_decision_dmn_handler,
)
from .tools.domains import (
    DOMAIN_TOOLS,
    create_child_domain_handler,
    create_domain_handler,
    get_domain_handler,
    list_domains_handler,
)
from .tools.instances import (
    INSTANCES_TOOLS,
    get_instance_compliance_handler,
    get_instance_handler,
    list_instances_handler,
)
from .tools.processes import (
    PROCESS_TOOLS,
    add_process_step_handler,
    create_process_handler,
    get_process_handler,
    list_processes_handler,
)
from .tools.products import (
    PRODUCT_TOOLS,
    add_product_requirement_handler,
    add_product_rule_handler,
    create_product_handler,
    delete_product_handler,
    get_product_handler,
    get_product_variants_handler,
    list_products_handler,
    remove_product_requirement_handler,
    remove_product_rule_handler,
    update_product_handler,
)
from .tools.requirements import (
    REQUIREMENTS_TOOLS,
    get_requirement_handler,
    list_requirements_handler,
    update_requirement_handler,
)
from .tools.rules import (
    RULES_TOOLS,
    create_rule_handler,
    get_rule_handler,
    list_rules_handler,
    update_rule_handler,
)
from .tools.services import (
    SERVICE_TOOLS,
    create_service_handler,
    get_service_handler,
    list_services_handler,
)
from .tools.validate import (
    VALIDATE_TOOLS,
    get_health_handler,
    validate_catalog_handler,
)
from .tools.workspace import (
    WORKSPACE_TOOLS,
    create_workspace_handler,
    get_current_workspace_handler,
    list_workspaces_handler,
)

app = Server("nomos-mcp")

ALL_TOOL_DEFS = (
    COSMOS_TOOLS
    + DOMAIN_TOOLS
    + SERVICE_TOOLS
    + DECISION_TOOLS
    + PROCESS_TOOLS
    + PRODUCT_TOOLS
    + RULES_TOOLS
    + REQUIREMENTS_TOOLS
    + BLUEPRINTS_TOOLS
    + INSTANCES_TOOLS
    + WORKSPACE_TOOLS
    + VALIDATE_TOOLS
)


@app.list_tools()
async def list_tools() -> list[Tool]:
    return [
        Tool(
            name=t["name"],
            description=t["description"],
            inputSchema=t["inputSchema"],
        )
        for t in ALL_TOOL_DEFS
    ]


@app.call_tool()
async def call_tool(name: str, arguments: dict) -> list[TextContent]:
    try:
        result = await _dispatch(name, arguments)
    except PermissionError as exc:
        result = f"Zugriff verweigert: {exc}"
    except Exception as exc:
        result = f"Fehler beim Ausführen von '{name}': {exc}"

    return [TextContent(type="text", text=str(result))]


async def _dispatch(name: str, args: dict) -> str:  # noqa: PLR0912, PLR0911
    # Cosmos
    if name == "cosmos_init":
        return await cosmos_init_handler(
            args["path"], args["id"], args["name"], args.get("owner", "")
        )
    if name == "cosmos_info":
        return await cosmos_info_handler()

    # Domains
    if name == "list_domains":
        return await list_domains_handler()
    if name == "get_domain":
        return await get_domain_handler(args["domain"])
    if name == "create_domain":
        return await create_domain_handler(args["dns_name"], args.get("owner", ""))
    if name == "create_child_domain":
        return await create_child_domain_handler(
            args["parent_domain"], args["segment"], args.get("owner", "")
        )

    # Services
    if name == "list_services":
        return await list_services_handler(args["domain"])
    if name == "get_service":
        return await get_service_handler(args["domain"], args["service"])
    if name == "create_service":
        return await create_service_handler(
            args["domain"],
            args["name"],
            args.get("owner", ""),
            args.get("summary", ""),
            args.get("capabilities"),
        )

    # Decisions
    if name == "list_decisions":
        return await list_decisions_handler(args["domain"])
    if name == "get_decision":
        return await get_decision_handler(args["domain"], args["decision_id"])
    if name == "create_decision":
        return await create_decision_handler(
            domain=args["domain"],
            name=args["name"],
            id=args.get("id", ""),
            number=args.get("number", ""),
            owner=args.get("owner", ""),
            summary=args.get("summary", ""),
            context=args.get("context", ""),
            status=args.get("status", "draft"),
            inputs=args.get("inputs"),
            outputs=args.get("outputs"),
        )
    if name == "update_decision_dmn":
        return await update_decision_dmn_handler(
            args["domain"], args["decision_id"], args["dmn_xml"]
        )
    if name == "get_decision_dmn":
        return await get_decision_dmn_handler(args["domain"], args["decision_id"])

    # Processes
    if name == "list_processes":
        return await list_processes_handler(args["product_id"])
    if name == "get_process":
        return await get_process_handler(args["process_id"])
    if name == "create_process":
        return await create_process_handler(
            args["product_id"],
            args["name"],
            args.get("summary", ""),
            args.get("id", ""),
        )
    if name == "add_process_step":
        return await add_process_step_handler(
            process_id=args["process_id"],
            name=args["name"],
            task_type=args["task_type"],
            service_ref=args.get("service_ref", ""),
            method=args.get("method", ""),
            decision_ref=args.get("decision_ref", ""),
            required=args.get("required", True),
            depends_on=args.get("depends_on"),
        )

    # Products (Python backend)
    if name == "list_products":
        return await list_products_handler()
    if name == "get_product":
        return await get_product_handler(args["product_id"])
    if name == "create_product":
        return await create_product_handler(args)
    if name == "update_product":
        return await update_product_handler(args["product_id"], args["updates"])
    if name == "delete_product":
        return await delete_product_handler(args["product_id"])
    if name == "add_product_requirement":
        return await add_product_requirement_handler(args["product_id"], args["requirement_id"])
    if name == "remove_product_requirement":
        return await remove_product_requirement_handler(args["product_id"], args["requirement_id"])
    if name == "add_product_rule":
        return await add_product_rule_handler(args["product_id"], args["rule_id"])
    if name == "remove_product_rule":
        return await remove_product_rule_handler(args["product_id"], args["rule_id"])
    if name == "get_product_variants":
        return await get_product_variants_handler(args["product_id"])

    # Rules (Python backend)
    if name == "list_rules":
        return await list_rules_handler()
    if name == "get_rule":
        return await get_rule_handler(args["rule_id"])
    if name == "create_rule":
        return await create_rule_handler(args)
    if name == "update_rule":
        return await update_rule_handler(args["rule_id"], args["updates"])

    # Requirements (Python backend)
    if name == "list_requirements":
        return await list_requirements_handler()
    if name == "get_requirement":
        return await get_requirement_handler(args["requirement_id"])
    if name == "update_requirement":
        return await update_requirement_handler(args["requirement_id"], args["updates"])

    # Blueprints (Python backend)
    if name == "list_blueprints":
        return await list_blueprints_handler()
    if name == "get_blueprint":
        return await get_blueprint_handler(args["blueprint_id"])

    # Instances (Python backend)
    if name == "list_instances":
        return await list_instances_handler()
    if name == "get_instance":
        return await get_instance_handler(args["instance_id"])
    if name == "get_instance_compliance":
        return await get_instance_compliance_handler(args["instance_id"])

    # Workspace
    if name == "create_workspace":
        return await create_workspace_handler(
            args["branch_name"],
            args.get("description", ""),
        )
    if name == "get_current_workspace":
        return await get_current_workspace_handler()
    if name == "list_workspaces":
        return await list_workspaces_handler()

    # Validate / Health
    if name == "validate_catalog":
        return await validate_catalog_handler(
            path=args.get("path", ""),
            fmt=args.get("format", "json"),
        )
    if name == "get_health":
        return await get_health_handler()

    return f"Unbekanntes Tool: '{name}'"


async def _main() -> None:
    async with stdio_server() as (read_stream, write_stream):
        await app.run(
            read_stream,
            write_stream,
            InitializationOptions(
                server_name="nomos-mcp",
                server_version="0.2.0",
                capabilities=app.get_capabilities(
                    notification_options=None,
                    experimental_capabilities={},
                ),
            ),
        )


def run() -> None:
    asyncio.run(_main())


if __name__ == "__main__":
    run()
