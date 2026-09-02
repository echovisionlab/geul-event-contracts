import crypto from "node:crypto";
import fs from "node:fs";
import path from "node:path";

import YAML from "yaml";

const namePattern = /^[a-z][a-z0-9_]*$/;

function readYaml(file) {
  return YAML.parse(fs.readFileSync(file, "utf8"));
}

function assertUniqueNames(values, label) {
  if (
    !Array.isArray(values) ||
    values.length === 0 ||
    new Set(values).size !== values.length ||
    values.some(
      (value) => typeof value !== "string" || !namePattern.test(value),
    )
  ) {
    throw new Error(`${label} must be a unique non-empty name list`);
  }
}

function validateDefinition(definitionName, capability, expressions) {
  if (!capability || typeof capability !== "object") {
    throw new Error(`${definitionName} must define capabilities`);
  }
  if (!capability.relations || typeof capability.relations !== "object") {
    throw new Error(`${definitionName} must define a relations map`);
  }
  if (!Array.isArray(capability.permissions)) {
    throw new Error(`${definitionName} must define a permissions list`);
  }
  if (!expressions || typeof expressions !== "object") {
    throw new Error(`${definitionName} must define permission expressions`);
  }

  for (const [relation, subjects] of Object.entries(capability.relations)) {
    if (
      !namePattern.test(relation) ||
      !Array.isArray(subjects) ||
      subjects.length === 0
    ) {
      throw new Error(
        `${definitionName}.${relation} must define non-empty subject references`,
      );
    }
    for (const subject of subjects) {
      if (
        typeof subject !== "string" ||
        !namePattern.test(subject.replace(/#member$/, ""))
      ) {
        throw new Error(
          `${definitionName}.${relation} has invalid subject reference ${subject}`,
        );
      }
    }
  }

  if (
    new Set(capability.permissions).size !== capability.permissions.length ||
    capability.permissions.some(
      (permission) =>
        typeof permission !== "string" || !namePattern.test(permission),
    )
  ) {
    throw new Error(`${definitionName} permissions must be unique names`);
  }
  const permissionSet = new Set(capability.permissions);
  const expressionNames = Object.keys(expressions);
  if (
    expressionNames.length !== permissionSet.size ||
    expressionNames.some((permission) => !permissionSet.has(permission))
  ) {
    throw new Error(
      `${definitionName} permission expressions must exactly match permissions`,
    );
  }
  for (const permission of capability.permissions) {
    if (
      typeof expressions[permission] !== "string" ||
      expressions[permission].trim() === ""
    ) {
      throw new Error(
        `${definitionName}.${permission} must define a non-empty expression`,
      );
    }
  }
}

function validateActions(module, definitions) {
  const actions = module.actions ?? {};
  if (!actions || typeof actions !== "object") {
    throw new Error(`${module.domain} actions must be a map`);
  }
  const definitionNames = Object.keys(definitions);
  if (
    JSON.stringify(Object.keys(actions).sort()) !==
    JSON.stringify([...definitionNames].sort())
  ) {
    throw new Error(
      `${module.domain} actions must explicitly cover every owned definition`,
    );
  }
  for (const [definitionName, mappings] of Object.entries(actions)) {
    const capability = definitions[definitionName];
    if (!capability || !mappings || typeof mappings !== "object") {
      throw new Error(
        `${module.domain} actions reference unknown definition ${definitionName}`,
      );
    }
    for (const [action, permission] of Object.entries(mappings)) {
      if (
        !namePattern.test(action) ||
        typeof permission !== "string" ||
        !capability.permissions.includes(permission)
      ) {
        throw new Error(
          `${module.domain}.${definitionName}.${action} has invalid permission ${permission}`,
        );
      }
    }
  }
  return actions;
}

function validateActionAliases(domain, definitions, actions) {
  const aliases = domain.action_aliases ?? {};
  if (!aliases || typeof aliases !== "object") {
    throw new Error(`${domain.domain} action_aliases must be a map`);
  }
  for (const [definitionName, mappings] of Object.entries(aliases)) {
    const capability = definitions[definitionName];
    if (!capability || !mappings || typeof mappings !== "object") {
      throw new Error(
        `${domain.domain} aliases reference unknown definition ${definitionName}`,
      );
    }
    for (const [action, permission] of Object.entries(mappings)) {
      if (
        !namePattern.test(action) ||
        typeof permission !== "string" ||
        !capability.permissions.includes(permission) ||
        capability.permissions.includes(action) ||
        Object.hasOwn(actions[definitionName] ?? {}, action)
      ) {
        throw new Error(
          `${domain.domain}.${definitionName}.${action} has invalid permission alias ${permission}`,
        );
      }
    }
  }
  return aliases;
}

function validatePlatformActions(module, definitions, platformPermissions) {
  const platformActions = module.platform_actions ?? {};
  if (!platformActions || typeof platformActions !== "object") {
    throw new Error(`${module.domain} platform_actions must be a map`);
  }
  for (const [definitionName, mappings] of Object.entries(platformActions)) {
    if (
      !definitions[definitionName] ||
      !mappings ||
      typeof mappings !== "object"
    ) {
      throw new Error(
        `${module.domain} platform actions reference unknown definition ${definitionName}`,
      );
    }
    for (const [action, permission] of Object.entries(mappings)) {
      if (
        !namePattern.test(action) ||
        typeof permission !== "string" ||
        !platformPermissions.includes(permission) ||
        Object.hasOwn(module.actions[definitionName] ?? {}, action) ||
        Object.hasOwn(module.action_aliases?.[definitionName] ?? {}, action)
      ) {
        throw new Error(
          `${module.domain}.${definitionName}.${action} has invalid platform permission ${permission}`,
        );
      }
    }
  }
  return platformActions;
}

function validateLookupPermissions(module, definitions) {
  const lookupPermissions = module.lookup_permissions ?? {};
  if (!lookupPermissions || typeof lookupPermissions !== "object") {
    throw new Error(`${module.domain} lookup_permissions must be a map`);
  }
  for (const [definitionName, permissions] of Object.entries(
    lookupPermissions,
  )) {
    const capability = definitions[definitionName];
    if (!capability) {
      throw new Error(
        `${module.domain} lookup permissions reference unknown definition ${definitionName}`,
      );
    }
    if (
      !Array.isArray(permissions) ||
      permissions.length === 0 ||
      new Set(permissions).size !== permissions.length
    ) {
      throw new Error(
        `${module.domain}.${definitionName} lookup permissions must be a unique non-empty list`,
      );
    }
    for (const permission of permissions) {
      if (
        typeof permission !== "string" ||
        !capability.permissions.includes(permission)
      ) {
        throw new Error(
          `${module.domain}.${definitionName} has invalid lookup permission ${permission}`,
        );
      }
    }
  }
  return lookupPermissions;
}

function validateSubjectLookups(core) {
  const subjectLookups = core.subject_lookups ?? {};
  if (
    !subjectLookups ||
    typeof subjectLookups !== "object" ||
    Object.keys(subjectLookups).length === 0
  ) {
    throw new Error("core subject_lookups must be a non-empty map");
  }
  const permissions = Object.values(subjectLookups);
  if (new Set(permissions).size !== permissions.length) {
    throw new Error("core subject_lookups permissions must be unique");
  }
  for (const [name, permission] of Object.entries(subjectLookups)) {
    if (
      !namePattern.test(name) ||
      typeof permission !== "string" ||
      !core.definitions.platform.permissions.includes(permission)
    ) {
      throw new Error(
        `core subject lookup ${name} has invalid platform permission ${permission}`,
      );
    }
  }
  return subjectLookups;
}

export function loadCatalog(root) {
  const catalogRoot = path.join(root, "config/spicedb/catalog");
  const manifest = readYaml(path.join(catalogRoot, "manifest.yaml"));
  const core = readYaml(path.join(catalogRoot, "core.yaml"));
  assertUniqueNames(manifest.modules, "catalog modules");
  assertUniqueNames(manifest.definition_order, "catalog definition_order");

  const domains = manifest.modules.map((moduleName) => {
    const file = path.join(catalogRoot, "domains", `${moduleName}.yaml`);
    const domain = readYaml(file);
    if (
      domain.domain !== moduleName ||
      typeof domain.description !== "string" ||
      domain.description.trim() === "" ||
      !domain.definitions ||
      typeof domain.definitions !== "object" ||
      !domain.permission_expressions ||
      typeof domain.permission_expressions !== "object"
    ) {
      throw new Error(`${moduleName} must be one described domain module`);
    }
    return { ...domain, file };
  });

  const discovered = fs
    .readdirSync(path.join(catalogRoot, "domains"), { withFileTypes: true })
    .filter((entry) => entry.isFile() && entry.name.endsWith(".yaml"))
    .map((entry) => entry.name.replace(/\.yaml$/, ""))
    .sort();
  if (
    JSON.stringify(discovered) !== JSON.stringify([...manifest.modules].sort())
  ) {
    throw new Error("manifest modules must exactly own every domain YAML file");
  }

  const definitionModules = [
    {
      domain: "core",
      description: "Account identity, role hierarchy, and platform policy core",
      definitions: core.definitions,
      permission_expressions: core.permission_expressions,
      actions: core.actions,
      action_aliases: {},
      platform_actions: {},
      lookup_permissions: {},
      file: path.join(catalogRoot, "core.yaml"),
    },
    ...domains,
  ];
  const unorderedDefinitions = {};
  const unorderedExpressions = {};
  const definitionOwners = new Map();
  const actions = {};
  const actionAliases = {};
  const platformActions = {};
  const lookupPermissions = {};
  for (const module of definitionModules) {
    const moduleDefinitions = Object.keys(module.definitions ?? {});
    if (
      moduleDefinitions.length === 0 ||
      Object.keys(module.permission_expressions ?? {}).length !==
        moduleDefinitions.length
    ) {
      throw new Error(`${module.domain} must own definitions and expressions`);
    }
    const moduleActions = validateActions(module, module.definitions);
    const aliases = validateActionAliases(
      module,
      module.definitions,
      moduleActions,
    );
    const modulePlatformActions = validatePlatformActions(
      module,
      module.definitions,
      core.definitions.platform.permissions,
    );
    const moduleLookupPermissions = validateLookupPermissions(
      module,
      module.definitions,
    );
    const apiNames = module.api_names ?? {};
    if (!apiNames || typeof apiNames !== "object") {
      throw new Error(`${module.domain} api_names must be a map`);
    }
    for (const [definitionName, apiName] of Object.entries(apiNames)) {
      if (
        !module.definitions[definitionName] ||
        typeof apiName !== "string" ||
        !/^[A-Z][A-Za-z0-9]*$/.test(apiName)
      ) {
        throw new Error(`${module.domain} has invalid API name ${apiName}`);
      }
    }
    for (const definitionName of moduleDefinitions) {
      if (
        !namePattern.test(definitionName) ||
        unorderedDefinitions[definitionName]
      ) {
        throw new Error(`duplicate or invalid definition ${definitionName}`);
      }
      validateDefinition(
        definitionName,
        module.definitions[definitionName],
        module.permission_expressions[definitionName],
      );
      unorderedDefinitions[definitionName] = module.definitions[definitionName];
      unorderedExpressions[definitionName] =
        module.permission_expressions[definitionName];
      definitionOwners.set(definitionName, module.domain);
      actions[definitionName] = moduleActions[definitionName];
      actionAliases[definitionName] = aliases[definitionName] ?? {};
      platformActions[definitionName] =
        modulePlatformActions[definitionName] ?? {};
      lookupPermissions[definitionName] =
        moduleLookupPermissions[definitionName] ?? [];
    }
  }

  const unorderedNames = Object.keys(unorderedDefinitions).sort();
  if (
    JSON.stringify(unorderedNames) !==
    JSON.stringify([...manifest.definition_order].sort())
  ) {
    throw new Error(
      "definition_order must contain every core and domain definition exactly once",
    );
  }

  const definitions = {};
  const permissionExpressions = {};
  for (const definitionName of manifest.definition_order) {
    definitions[definitionName] = unorderedDefinitions[definitionName];
    permissionExpressions[definitionName] =
      unorderedExpressions[definitionName];
  }

  if (
    core.subject_type !== "account_identity" ||
    core.platform?.type !== "platform" ||
    core.platform?.id !== "global" ||
    definitionOwners.get("account_identity") !== "core" ||
    definitionOwners.get("role") !== "core" ||
    definitionOwners.get("platform") !== "core"
  ) {
    throw new Error(
      "core must own account_identity, role, and platform:global",
    );
  }
  assertUniqueNames(core.roles, "catalog roles");
  const subjectLookups = validateSubjectLookups(core);
  if (
    !Array.isArray(core.role_inheritance) ||
    core.role_inheritance.length !== core.roles.length - 1
  ) {
    throw new Error(
      "catalog role inheritance must contain one edge per non-root role",
    );
  }
  for (const edge of core.role_inheritance) {
    if (
      !core.roles.includes(edge.role) ||
      !core.roles.includes(edge.includes) ||
      edge.role === edge.includes
    ) {
      throw new Error(`invalid role inheritance edge: ${JSON.stringify(edge)}`);
    }
  }

  for (const [definitionName, capability] of Object.entries(definitions)) {
    for (const subjects of Object.values(capability.relations)) {
      for (const subject of subjects) {
        const subjectDefinition = subject.replace(/#member$/, "");
        if (!definitions[subjectDefinition]) {
          throw new Error(
            `${definitionName} references unknown subject ${subject}`,
          );
        }
      }
    }
  }

  const source = {
    subject_type: core.subject_type,
    platform: core.platform,
    roles: core.roles,
    role_inheritance: core.role_inheritance,
    definitions,
    permission_expressions: permissionExpressions,
  };
  const fingerprint = crypto
    .createHash("sha256")
    .update(JSON.stringify(source))
    .digest("hex");
  const resourceTypes = manifest.definition_order.filter(
    (name) => name !== "account_identity" && name !== "role",
  );

  return {
    source,
    fingerprint,
    definitions,
    permissionExpressions,
    definitionNames: manifest.definition_order,
    definitionOwners,
    definitionModules,
    domains,
    actions,
    actionAliases,
    platformActions,
    lookupPermissions,
    apiNames: Object.fromEntries(
      definitionModules.flatMap((module) =>
        Object.entries(module.api_names ?? {}),
      ),
    ),
    resourceTypes,
    roles: core.roles,
    roleInheritance: core.role_inheritance,
    subjectLookups,
    platform: core.platform,
    subjectType: core.subject_type,
  };
}

export function pascal(value) {
  return value
    .split(/[_#]/)
    .filter(Boolean)
    .map((part) => part[0].toUpperCase() + part.slice(1))
    .join("");
}
