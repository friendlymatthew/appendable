import { FieldType } from "../db/database";
import { FileFormat, readFileMeta, readIndexMeta } from "../file/meta";
import { readBinaryFile } from "./test-util";

describe("test file parsing", () => {
  let fileMetaBuffer: Uint8Array;
  let indexMetaBuffer: Uint8Array;

  beforeAll(async () => {
    fileMetaBuffer = await readBinaryFile("filemeta.bin");
    indexMetaBuffer = await readBinaryFile("indexmeta.bin");
  });

  it("should read the file meta", async () => {
    const fileMeta = await readFileMeta(fileMetaBuffer.buffer);
    expect(fileMeta.format).toEqual(FileFormat.CSV);
    expect(fileMeta.version).toEqual(1);
    expect(fileMeta.readOffset).toEqual(4096n);
    expect(fileMeta.entries).toEqual(34);
  });

  it("should read the index meta", async () => {
    const indexMeta = await readIndexMeta(indexMetaBuffer.buffer);
    expect(indexMeta.width).toEqual(2);
    expect(indexMeta.fieldName).toEqual("howdydo");
    expect(indexMeta.fieldType).toEqual(FieldType.Boolean);
    expect(indexMeta.totalFieldValueLength).toEqual(773424601);
  });
});
