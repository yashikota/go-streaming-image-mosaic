package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"os"
)

// モザイク処理に必要な情報を保持する構造体
type MosaicProcessor struct {
	img          *image.NRGBA // 元画像
	mosaicWidth  int          // モザイクタイルの幅
	mosaicHeight int          // モザイクタイルの高さ
	buffer       *image.NRGBA // 一部画像を一時的に保持するバッファ
	mosaicOffset int          // 処理中の画像のオフセット
}

// インスタンスを生成
func NewMosaicProcessor(img *image.NRGBA, mosaicWidth, mosaicHeight int) *MosaicProcessor {
	// バッファを生成 (画像の幅 × モザイクの高さ)
	buffer := image.NewNRGBA(image.Rect(0, 0, img.Bounds().Max.X, mosaicHeight))
	return &MosaicProcessor{
		img:          img,
		mosaicWidth:  mosaicWidth,
		mosaicHeight: mosaicHeight,
		buffer:       buffer,
		mosaicOffset: 0,
	}
}

// モザイク処理を実行し、処理後の画像を返却
func (mp *MosaicProcessor) Process() *image.NRGBA {
	bounds := mp.img.Bounds() // 元画像の範囲を取得

	// 出力画像を生成 (元画像と同じサイズ)
	output := image.NewNRGBA(bounds)

	// 画像をモザイクタイルの高さ単位で処理
	for mp.mosaicOffset < bounds.Max.Y {
		// バッファに画像の一部を読み込む
		mp.readToBuffer()

		// バッファ内のデータをモザイク処理
		mp.applyMosaicToBuffer()

		// 処理済みのデータを出力画像にコピー
		mp.copyBufferToOutput(output)

		// 次の処理部分へオフセットを更新
		mp.mosaicOffset += mp.mosaicHeight

		fmt.Println(mp.mosaicOffset)
	}

	return output
}

// バッファに画像の一部を読み込む
func (mp *MosaicProcessor) readToBuffer() {
	// バッファに、元の画像から指定範囲をコピー
	draw.Draw(mp.buffer, mp.buffer.Bounds(), mp.img, image.Point{0, mp.mosaicOffset}, draw.Src)
}

// バッファ内のデータをモザイク処理
func (mp *MosaicProcessor) applyMosaicToBuffer() {
	bounds := mp.buffer.Bounds() // バッファの範囲を取得

	// バッファをモザイクタイル単位で処理
	for y := bounds.Min.Y; y < bounds.Max.Y; y += mp.mosaicHeight {
		for x := bounds.Min.X; x < bounds.Max.X; x += mp.mosaicWidth {
			// モザイクタイルの平均色を計算
			avgColor := mp.averageColor(x, y, mp.mosaicWidth, mp.mosaicHeight)

			// モザイクタイルを平均色で塗りつぶす
			for dy := 0; dy < mp.mosaicHeight && y+dy < bounds.Max.Y; dy++ {
				for dx := 0; dx < mp.mosaicWidth && x+dx < bounds.Max.X; dx++ {
					mp.buffer.Set(x+dx, y+dy, avgColor)
				}
			}
		}
	}
}

// 処理済みのデータを出力画像にコピー
func (mp *MosaicProcessor) copyBufferToOutput(output *image.NRGBA) {
	// コピーする高さを計算
	copyHeight := min(mp.mosaicHeight, output.Bounds().Max.Y-mp.mosaicOffset)

	// 出力画像に、バッファから指定範囲をコピー
	draw.Draw(output, image.Rect(0, mp.mosaicOffset, output.Bounds().Max.X, mp.mosaicOffset+copyHeight),
		mp.buffer, image.Point{0, 0}, draw.Src)
}

func (mp *MosaicProcessor) averageColor(x, y, width, height int) color.Color {
	var r, g, b, a uint32
	var count uint32
	bounds := mp.buffer.Bounds()
	// 指定範囲の画素の平均色を計算
	for dy := 0; dy < height && y+dy < bounds.Max.Y; dy++ {
		for dx := 0; dx < width && x+dx < bounds.Max.X; dx++ {
			pr, pg, pb, pa := mp.buffer.At(x+dx, y+dy).RGBA()
			r += pr
			g += pg
			b += pb
			a += pa
			count++
		}
	}
	if count == 0 {
		return color.NRGBA{0, 0, 0, 255}
	}
	return color.NRGBA{
		R: uint8(r / count >> 8),
		G: uint8(g / count >> 8),
		B: uint8(b / count >> 8),
		A: uint8(a / count >> 8),
	}
}

func convertToNRGBA(img image.Image) *image.NRGBA {
	bounds := img.Bounds()
	nrgba := image.NewNRGBA(bounds)
	draw.Draw(nrgba, bounds, img, image.Point{}, draw.Src)
	return nrgba
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	file, err := os.Open("test.jpg")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		panic(err)
	}

	nrgbaImg := convertToNRGBA(img)

	mosaicWidth := 100
	mosaicHeight := 100

	processor := NewMosaicProcessor(nrgbaImg, mosaicWidth, mosaicHeight)
	output := processor.Process()

	outFile, err := os.Create("result.jpg")
	if err != nil {
		panic(err)
	}
	defer outFile.Close()

	jpeg.Encode(outFile, output, nil)
}
