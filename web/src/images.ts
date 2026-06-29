// prepareImageUpload normalizes a picked file for upload. iPhone HEIC/HEIF is
// converted to JPEG client-side so the server can stay CGO-free; every other
// format is passed through untouched. Returns the blob to upload plus a filename
// whose extension matches the (possibly converted) bytes.
export async function prepareImageUpload(file: File): Promise<{ blob: Blob; name: string }> {
  if (/heic|heif/i.test(file.type) || /\.hei[cf]$/i.test(file.name)) {
    const heic2any = (await import('heic2any')).default
    const blob = (await heic2any({ blob: file, toType: 'image/jpeg', quality: 0.9 })) as Blob
    return { blob, name: file.name.replace(/\.\w+$/, '.jpg') }
  }
  return { blob: file, name: file.name }
}
