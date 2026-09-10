import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import YAML from "yaml";

const repoRoot = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../..",
);
const protoRoot = path.join(repoRoot, "proto");
const asyncapiPath = path.join(repoRoot, "asyncapi/asyncapi.yaml");
const rpcInventoryPath = path.join(repoRoot, "docs/rpc-inventory.json");
const syncRpcInventory = process.argv.includes("--sync-rpc-inventory");

const failures = [];

function walk(root, accept, skipDirectory = () => false) {
  const files = [];
  const visit = (directory) => {
    for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
      const fullPath = path.join(directory, entry.name);
      if (entry.isDirectory()) {
        if (!skipDirectory(entry.name, fullPath)) visit(fullPath);
      } else if (entry.isFile() && accept(fullPath)) {
        files.push(fullPath);
      }
    }
  };
  visit(root);
  return files.sort();
}

function stripProtoComments(source) {
  return source.replace(/\/\*[\s\S]*?\*\//g, "").replace(/\/\/.*$/gm, "");
}

function extractNamedBlocks(source, keyword) {
  const blocks = [];
  const pattern = new RegExp(`\\b${keyword}\\s+([\\w.]+)\\s*\\{`, "g");
  let match;
  while ((match = pattern.exec(source)) !== null) {
    const open = source.indexOf("{", match.index);
    let depth = 1;
    let cursor = open + 1;
    while (cursor < source.length && depth > 0) {
      if (source[cursor] === "{") depth += 1;
      if (source[cursor] === "}") depth -= 1;
      cursor += 1;
    }
    if (depth !== 0) throw new Error(`unterminated ${keyword} ${match[1]}`);
    blocks.push({ name: match[1], body: source.slice(open + 1, cursor - 1) });
    pattern.lastIndex = cursor;
  }
  return blocks;
}

function resolveMessage(typeName, packageName, messages, simpleNames) {
  const normalized = typeName.replace(/^\./, "");
  if (messages.has(normalized)) return normalized;
  const localName = `${packageName}.${normalized}`;
  if (messages.has(localName)) return localName;
  const candidates = simpleNames.get(normalized) ?? [];
  return candidates.length === 1 ? candidates[0] : undefined;
}

function parseContracts() {
  const protoFiles = walk(protoRoot, (file) => file.endsWith(".proto"));
  const messages = new Map();
  const simpleNames = new Map();
  const services = [];
  const extensionRoots = [];
  const outOfBandRoots = [];

  for (const file of protoFiles) {
    const rawSource = fs.readFileSync(file, "utf8");
    const source = stripProtoComments(rawSource);
    const packageName = source.match(/\bpackage\s+([\w.]+)\s*;/)?.[1];
    if (!packageName) throw new Error(`missing package declaration: ${file}`);

    const outOfBandRootPattern =
      /\/\/\s*geul-contract-root:\s*[a-z0-9-]+\s*\r?\n\s*message\s+(\w+)\s*\{/g;
    let outOfBandRootMatch;
    while (
      (outOfBandRootMatch = outOfBandRootPattern.exec(rawSource)) !== null
    ) {
      outOfBandRoots.push({
        packageName,
        typeName: outOfBandRootMatch[1],
      });
    }

    for (const block of extractNamedBlocks(source, "message")) {
      const fullName = `${packageName}.${block.name}`;
      messages.set(fullName, { ...block, packageName, file });
      const names = simpleNames.get(block.name) ?? [];
      names.push(fullName);
      simpleNames.set(block.name, names);
    }

    for (const block of extractNamedBlocks(source, "service")) {
      const rpcs = [];
      const rpcPattern =
        /\brpc\s+(\w+)\s*\(\s*(?:stream\s+)?([\w.]+)\s*\)\s*returns\s*\(\s*(?:stream\s+)?([\w.]+)\s*\)/g;
      let rpcMatch;
      while ((rpcMatch = rpcPattern.exec(block.body)) !== null) {
        rpcs.push({
          name: rpcMatch[1],
          request: rpcMatch[2],
          response: rpcMatch[3],
        });
      }
      services.push({ packageName, name: block.name, rpcs, file });
    }

    for (const block of extractNamedBlocks(source, "extend")) {
      const fieldPattern =
        /(?:optional\s+|repeated\s+)?([.\w]+)\s+\w+\s*=\s*\d+/g;
      let fieldMatch;
      while ((fieldMatch = fieldPattern.exec(block.body)) !== null) {
        extensionRoots.push({ packageName, typeName: fieldMatch[1] });
      }
    }
  }

  return { messages, simpleNames, services, extensionRoots, outOfBandRoots };
}

function collectMessageEdges(message, contracts) {
  const edges = new Set();
  const fieldPattern =
    /(?:^|\n)\s*(?:optional\s+|repeated\s+)?(map\s*<[^>]+>|[.\w]+)\s+\w+\s*=\s*\d+/g;
  let fieldMatch;
  while ((fieldMatch = fieldPattern.exec(message.body)) !== null) {
    let typeName = fieldMatch[1];
    if (typeName.startsWith("map")) {
      typeName = typeName.slice(typeName.indexOf(",") + 1, -1).trim();
    }
    const resolved = resolveMessage(
      typeName,
      message.packageName,
      contracts.messages,
      contracts.simpleNames,
    );
    if (resolved) edges.add(resolved);
  }
  return edges;
}

function collectRefs(value, refs = []) {
  if (Array.isArray(value)) {
    for (const item of value) collectRefs(item, refs);
  } else if (value && typeof value === "object") {
    for (const [key, item] of Object.entries(value)) {
      if (key === "$ref" && typeof item === "string") refs.push(item);
      else collectRefs(item, refs);
    }
  }
  return refs;
}

function auditContractGraph(contracts) {
  const asyncapi = YAML.parse(fs.readFileSync(asyncapiPath, "utf8"));
  const components = asyncapi.components?.messages ?? {};
  const channelRefs = collectRefs(asyncapi.channels ?? {}).filter((ref) =>
    ref.startsWith("#/components/messages/"),
  );
  const referencedComponents = new Set(
    channelRefs.map((ref) => ref.split("/").at(-1)),
  );

  for (const name of Object.keys(components)) {
    if (!referencedComponents.has(name)) {
      failures.push(
        `AsyncAPI message component is not used by a channel: ${name}`,
      );
    }
  }

  const roots = new Set();
  for (const service of contracts.services) {
    for (const rpc of service.rpcs) {
      for (const typeName of [rpc.request, rpc.response]) {
        const resolved = resolveMessage(
          typeName,
          service.packageName,
          contracts.messages,
          contracts.simpleNames,
        );
        if (!resolved)
          failures.push(
            `RPC references unknown message: ${service.packageName}.${service.name}.${rpc.name} -> ${typeName}`,
          );
        else roots.add(resolved);
      }
    }
  }

  for (const { packageName, typeName } of contracts.extensionRoots) {
    const resolved = resolveMessage(
      typeName,
      packageName,
      contracts.messages,
      contracts.simpleNames,
    );
    if (resolved) roots.add(resolved);
  }

  for (const { packageName, typeName } of contracts.outOfBandRoots) {
    const resolved = resolveMessage(
      typeName,
      packageName,
      contracts.messages,
      contracts.simpleNames,
    );
    if (!resolved) {
      failures.push(
        `out-of-band contract root references unknown message: ${packageName}.${typeName}`,
      );
    } else {
      roots.add(resolved);
    }
  }

  for (const name of referencedComponents) {
    if (!(name in components)) {
      failures.push(`AsyncAPI channel references missing component: ${name}`);
      continue;
    }
    const resolved = resolveMessage(
      name,
      "",
      contracts.messages,
      contracts.simpleNames,
    );
    if (!resolved)
      failures.push(
        `AsyncAPI component has no unique protobuf message: ${name}`,
      );
    else roots.add(resolved);
  }

  const reachable = new Set();
  const pending = [...roots];
  while (pending.length > 0) {
    const current = pending.pop();
    if (!current || reachable.has(current)) continue;
    reachable.add(current);
    const message = contracts.messages.get(current);
    if (!message) continue;
    for (const edge of collectMessageEdges(message, contracts))
      pending.push(edge);
  }

  for (const [name, message] of contracts.messages) {
    if (!reachable.has(name)) {
      failures.push(
        `protobuf message is unreachable from every RPC/event root: ${path.relative(repoRoot, message.file)}:${message.name}`,
      );
    }
  }

  return {
    asyncapi,
    componentCount: Object.keys(components).length,
    messageCount: contracts.messages.size,
    reachableCount: reachable.size,
  };
}

function contractRpcServices(contracts) {
  return contracts.services
    .filter((service) =>
      ["intra", "manage", "open"].includes(service.packageName.split(".")[1]),
    )
    .map((service) => ({
      name: `${service.packageName}.${service.name}`,
      procedures: service.rpcs.map((rpc) => rpc.name).sort(),
    }))
    .sort((left, right) => left.name.localeCompare(right.name));
}

function sortedUnique(values) {
  return [...new Set(values)].sort();
}

function syncRpcInventoryFile(contracts) {
  const existing = fs.existsSync(rpcInventoryPath)
    ? JSON.parse(fs.readFileSync(rpcInventoryPath, "utf8"))
    : {
        schemaVersion: 2,
        consumerAuditStatus: "in_progress",
        services: [],
      };
  const existingServices = new Map(
    (existing.services ?? []).map((service) => [service.service, service]),
  );

  const services = contractRpcServices(contracts).map((contractService) => {
    const prior = existingServices.get(contractService.name) ?? {};
    const priorProcedures = new Map(
      (prior.procedures ?? []).map((procedure) => [procedure.name, procedure]),
    );

    return {
      service: contractService.name,
      procedures: contractService.procedures.map((name) => {
        const previous = priorProcedures.get(name);
        return {
          name,
          consumerStatus:
            previous?.consumerStatus === "verified" ? "verified" : "pending",
          consumers: previous?.consumers ?? [],
          consumerTests: previous?.consumerTests ?? [],
          providerTests: previous?.providerTests ?? [],
        };
      }),
    };
  });

  const inventory = {
    schemaVersion: 2,
    consumerAuditStatus:
      existing.schemaVersion === 2 &&
      existing.consumerAuditStatus === "complete"
        ? "complete"
        : "in_progress",
    services,
  };
  fs.writeFileSync(rpcInventoryPath, `${JSON.stringify(inventory, null, 2)}\n`);
}

function auditRpcInventory(contracts) {
  if (!fs.existsSync(rpcInventoryPath)) {
    failures.push(`missing RPC inventory: ${rpcInventoryPath}`);
    return {
      serviceCount: 0,
      procedureCount: 0,
      consumerVerifiedCount: 0,
      consumerPendingCount: 0,
      consumerAuditStatus: "missing",
    };
  }

  const inventory = JSON.parse(fs.readFileSync(rpcInventoryPath, "utf8"));
  if (inventory.schemaVersion !== 2) {
    failures.push(
      `unsupported RPC inventory schema version: ${inventory.schemaVersion}`,
    );
  }
  if (!["in_progress", "complete"].includes(inventory.consumerAuditStatus)) {
    failures.push(
      `invalid consumer audit status: ${inventory.consumerAuditStatus}`,
    );
  }

  const contractServices = new Map(
    contractRpcServices(contracts).map((service) => [service.name, service]),
  );
  const inventoryServices = new Map();
  let procedureCount = 0;
  let consumerVerifiedCount = 0;
  let consumerPendingCount = 0;

  for (const service of inventory.services ?? []) {
    if (inventoryServices.has(service.service)) {
      failures.push(`duplicate RPC inventory service: ${service.service}`);
      continue;
    }
    inventoryServices.set(service.service, service);

    const contractService = contractServices.get(service.service);
    if (!contractService) {
      failures.push(`stale RPC inventory service: ${service.service}`);
      continue;
    }

    const expected = contractService.procedures;
    const procedures = service.procedures ?? [];
    const actual = sortedUnique(procedures.map((procedure) => procedure.name));
    if (JSON.stringify(actual) !== JSON.stringify(expected)) {
      failures.push(`RPC inventory differs from protobuf: ${service.service}`);
    }
    if (actual.length !== procedures.length) {
      failures.push(`duplicate RPC inventory procedure: ${service.service}`);
    }

    for (const procedure of procedures) {
      const fullName = `${service.service}/${procedure.name}`;
      if (!["pending", "verified"].includes(procedure.consumerStatus)) {
        failures.push(
          `invalid RPC consumer status: ${fullName} -> ${procedure.consumerStatus}`,
        );
        continue;
      }
      if (procedure.consumerStatus === "pending") {
        consumerPendingCount += 1;
        continue;
      }

      consumerVerifiedCount += 1;
      if ((procedure.consumers ?? []).length === 0) {
        failures.push(`verified RPC lacks consumer evidence: ${fullName}`);
      }
      for (const [index, consumer] of (procedure.consumers ?? []).entries()) {
        for (const field of [
          "repository",
          "file",
          "symbol",
          "call",
          "entrypoint",
        ]) {
          if (
            typeof consumer[field] !== "string" ||
            consumer[field].trim() === ""
          ) {
            failures.push(
              `RPC consumer evidence lacks ${field}: ${fullName} consumer ${index + 1}`,
            );
          }
        }
      }
    }

    procedureCount += expected.length;
  }

  for (const serviceName of contractServices.keys()) {
    if (!inventoryServices.has(serviceName)) {
      failures.push(
        `protobuf service is missing from RPC inventory: ${serviceName}`,
      );
    }
  }
  if (
    inventory.consumerAuditStatus === "complete" &&
    consumerPendingCount > 0
  ) {
    failures.push(
      `consumer audit is marked complete with ${consumerPendingCount} pending RPCs`,
    );
  }

  return {
    serviceCount: contractServices.size,
    procedureCount,
    consumerVerifiedCount,
    consumerPendingCount,
    consumerAuditStatus: inventory.consumerAuditStatus,
  };
}

function auditManagedAssetScalars(contracts) {
  const canonicalMediaFile = "proto/api/common/v1/media.proto";
  const managedAssetToken =
    /(?:^|_)(?:image|avatar|logo|favicon|loader|poster|artwork|thumbnail|featured_image|waveform|spectrogram|audio|hls|og)(?:_|$)/;
  const managedScalarName = /(?:_url|_urls|_key|_keys)$/;
  const exactManagedNames = new Set([
    "image",
    "avatar",
    "logo",
    "poster",
    "artwork",
    "thumbnail",
    "featured_image",
    "credited_as_image",
  ]);
  let fieldCount = 0;

  for (const message of contracts.messages.values()) {
    const relativeFile = path
      .relative(repoRoot, message.file)
      .split(path.sep)
      .join("/");
    if (
      !relativeFile.startsWith("proto/api/") ||
      relativeFile === canonicalMediaFile
    ) {
      continue;
    }

    const fieldPattern =
      /(?:^|\n)\s*(?:optional\s+|repeated\s+)?string\s+(\w+)\s*=\s*\d+/g;
    let fieldMatch;
    while ((fieldMatch = fieldPattern.exec(message.body)) !== null) {
      fieldCount += 1;
      const fieldName = fieldMatch[1];
      const isManagedScalar =
        exactManagedNames.has(fieldName) ||
        (managedAssetToken.test(fieldName) &&
          managedScalarName.test(fieldName));
      if (isManagedScalar) {
        failures.push(
          `managed asset field must use a canonical media contract: ${message.packageName}.${message.name}.${fieldName}`,
        );
      }
    }
  }

  return fieldCount;
}

const contracts = parseContracts();
if (syncRpcInventory) {
  syncRpcInventoryFile(contracts);
  console.log(`Synchronized RPC inventory: ${rpcInventoryPath}`);
}
const contractAudit = auditContractGraph(contracts);
const rpcInventoryAudit = auditRpcInventory(contracts);
auditManagedAssetScalars(contracts);
console.log(
  `Contract graph: ${contractAudit.reachableCount}/${contractAudit.messageCount} protobuf messages reachable; ${contractAudit.componentCount} AsyncAPI message components linked`,
);
console.log(
  `RPC consumer audit (${rpcInventoryAudit.consumerAuditStatus}): ${rpcInventoryAudit.consumerVerifiedCount}/${rpcInventoryAudit.procedureCount} manually verified; ${rpcInventoryAudit.consumerPendingCount} pending across ${rpcInventoryAudit.serviceCount} services`,
);

if (failures.length > 0) {
  for (const failure of failures) console.error(`- ${failure}`);
  process.exit(1);
}
