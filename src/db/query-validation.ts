import { IndexHeader } from "../file/meta";
import { FieldType, fieldTypeToString } from "./database";
import {
  OrderBy,
  Schema,
  Query,
  SelectField,
  WhereNode,
  Search,
} from "./query-lang";

function checkType(headerType: number[], queryType: FieldType): boolean {
  return headerType.includes(queryType);
}

function validateWhere<T extends Schema>(
  where: WhereNode<T>[] | undefined,
  headers: IndexHeader[],
): void {
  if (!where || !Array.isArray(where) || where.length === 0) {
    throw new Error("Missing 'where' clause.");
  }

  for (const whereNode of where) {
    if (!["<", "<=", "==", ">=", ">"].includes(whereNode.operation)) {
      throw new Error("Invalid operation in 'where' clause.");
    }

    if (typeof whereNode.key !== "string") {
      throw new Error("'key' in 'where' clause must be a string.");
    }

    const header = headers.find((h) => h.fieldName === whereNode.key);

    if (!header) {
      throw new Error(
        `key: ${whereNode.key} in 'where' clause does not exist in dataset.`,
      );
    }

    if (typeof whereNode.value === "undefined") {
      throw new Error("'value' in 'where' clause is missing.");
    }

    const headerType = header.fieldTypes;

    if (whereNode.value === null) {
      if (!checkType(headerType, FieldType.Null)) {
        throw new Error(
          `null type not included in ${whereNode.key}'s header types.`,
        );
      }
    } else {
      switch (typeof whereNode.value) {
        case "bigint":
        case "number":
          if (
            !checkType(headerType, FieldType.Float64) &&
            !checkType(headerType, FieldType.Uint64) &&
            !checkType(headerType, FieldType.Int64)
          ) {
            throw new Error(
              `number type not included in ${whereNode.key}'s header types.`,
            );
          }
          break;

        case "string":
          if (!checkType(headerType, FieldType.String)) {
            throw new Error(
              `string type not included in ${whereNode.key}'s header types`,
            );
          }
          break;

        case "boolean":
          if (!checkType(headerType, FieldType.Boolean)) {
            throw new Error(
              `boolean type not included in ${whereNode.key}'s header types`,
            );
          }
          break;

        default:
          throw new Error(
            `unrecognized type: ${typeof whereNode.value} not included in ${whereNode.key}'s header types`,
          );
      }
    }
  }
}

function validateOrderBy<T extends Schema>(
  orderBy: OrderBy<T>[] | undefined,
  whereKey: string,
): void {
  if (orderBy) {
    if (!Array.isArray(orderBy) || orderBy.length === 0) {
      throw new Error("Invalid 'orderBy' clause.");
    }

    // Note: currently we only support one orderBy and it must be the where clause. When we add composite indexes and complex querying, refactor.
    const orderByObj = orderBy[0];

    if (!["ASC", "DESC"].includes(orderByObj.direction)) {
      throw new Error("Invalid direction in `orderBy`.");
    }

    if (orderByObj.key !== whereKey) {
      throw new Error("'key' in `orderBy` must match `key` in `where` clause");
    }
  }
}

function validateSelect<T extends Schema>(
  select: SelectField<T>[] | undefined,
  headers: IndexHeader[],
): void {
  if (select) {
    if (!Array.isArray(select)) {
      throw new Error(`select is not an array: ${select}`);
    }

    if (select.length <= 0) {
      throw new Error(`select clause is empty: ${select}`);
    }

    let hset = new Set<string>();
    headers.map((h) => hset.add(h.fieldName));

    select.map((s) => {
      if (!hset.has(s as string)) {
        throw new Error(
          `${s as string} is not included in the field name headers`,
        );
      }
    });
  }
}

export function validateSearch<T extends Schema>(
  search: Search<T>,
  headers: IndexHeader[],
) {
  if (!search.config) {
    search.config = {
      minGram: 1,
      maxGram: 2,
    };
  }
  const { config } = search;
  let { minGram, maxGram } = config;

  const fh = headers.find((h) => h.fieldName === search.key);

  if (!fh) {
    throw new Error(
      `Unable to find index header for key: ${search.key as string}`,
    );
  }

  let gset = new Set([FieldType.Unigram, FieldType.Bigram, FieldType.Trigram]);
  const { fieldTypes } = fh;
  fieldTypes.forEach((ft) => (gset.has(ft) ? gset.delete(ft) : {}));

  if (gset.size != 0) {
    throw new Error(
      `Unable to find valid ngram field types: ${[...gset.keys()].map((f) => fieldTypeToString(f))} for index header: ${search.key as string}.`,
    );
  }

  if (maxGram > 3 || minGram > 3) {
    throw new Error(
      `Invalid gram length configuration. ${config.minGram} and ${config.maxGram} cannot be greater than 3.`,
    );
  }

  if (minGram < 1 || maxGram < 1) {
    throw new Error(
      `Invalid gram length configuration. ${config.minGram} and ${config.maxGram} cannot be less than 3.`,
    );
  }

  if (minGram > maxGram) {
    throw new Error(
      `Invalid gram length configuration: minGram ${config.minGram} cannot be greater than maxGram ${config.maxGram}.`,
    );
  }
}

export function validateQuery<T extends Schema>(
  query: Query<T>,
  headers: IndexHeader[],
): void {
  if (query.search) {
    validateSearch(query.search, headers);
  }

  if (query.where) {
    validateWhere(query.where, headers);
    validateOrderBy(query.orderBy, query.where![0].key as string);
    validateSelect(query.select, headers);
  }
}
