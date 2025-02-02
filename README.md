# AquesTalk TTS API サーバー (OpenAI Compatible)

このリポジトリは、AquesTalk (旧ライセンス) の動的ライブラリ (DLL) を呼び出して、日本語音声合成を提供するHTTP APIサーバーを実装したものです。APIはOpenAIの音声合成APIに互換しており、主に「ゆっくりボイス」の音声を生成します。

## 特徴

- **超軽量な日本語音声合成**: AquesTalk (旧ライセンス) を利用した日本語音声合成。
- **営利利用可**: 旧ライセンス版を使用することで、営利事業であってもライセンス使用料を支払う必要はありません。
- **OpenAI互換API**: OpenAIクライアントライブラリで利用可能な音声合成API。

## 使用上の注意

- **プラットフォーム制限**: AquesTalkはWindowsプラットフォーム専用で、提供されるDLLおよび実行ファイルもWindows専用です。Wine等を用いて他のプラットフォームで使用することはライセンス違反となります。
- **ライセンス**: 本プロジェクトはGPL 3.0ライセンスの下で提供されます。AquesTalkの使用には`AqLicense.txt`を読み、ライセンスに従ってください。

## インストール方法

1. **リリースからEXEをダウンロード**  
   リリースページからWindows用の実行ファイル (`.exe`) をダウンロードしてください。

2. **EXEを実行**  
   ダウンロードした実行ファイルを起動すると、APIサーバーがポート8080で起動します。

3. **APIの利用**  
   APIサーバーが起動したら、OpenAIクライアントライブラリを使用して音声合成APIを呼び出すことができます。

## APIエンドポイント

- **POST** `/v1/audio/speech`
  
  音声合成を行うエンドポイントです。

### リクエスト例 (curl)

```
curl http://localhost:8080/v1/audio/speech \
  -H "Content-Type: application/json" \
  -d '{
    "model": "tts-1",
    "input": "おはようございます。",
    "voice": "f1",
    "response_format": "wav",
    "speed": 1.0
  }' \
  --output speech.wav
```

### パラメータ

- `model` (string): 固定値 `tts-1`。`tts-1-hd` は使用できません。
- `voice` (string): 使用する音声の種類。`dvd`, `f1`, `f2`, `imd1`, `jgr`, `m1`, `m2`, `r1` のいずれか。
- `input` (string): 合成するテキスト。漢字等には対応していません。
- `response_format` (string): 固定値 `wav`。それ以外はエラーとなります。
- `speed` (float): 音声の速度。0.5から3.0の間の値で指定します。

## サンプルコード (Python)

以下のサンプルコードでは、OpenAIクライアントライブラリを使ってAPIを呼び出します。

```python
from openai import OpenAI

client = OpenAI(api_key="a", base_url="http://localhost:8080/v1")
response = client.audio.speech.create(
    model="tts-1",
    voice="f1",
    input="おはようございます。",
)
```

## 免責事項

- 本プロジェクトは、AquesTalkのライセンスに基づき、Windowsプラットフォームでのみ使用可能です。他のプラットフォームで使用した場合、ライセンス違反となり、その責任はユーザーにあります。
- 本リポジトリの開発者は、ライセンス違反に対して一切の責任を負いません。

## ライセンス

本プロジェクトはGPL 3.0ライセンスの下で提供されています。AquesTalk (旧ライセンス版) を使用する際には、[AqLicense.txt](AqLicense.txt)を必ず確認し、ライセンスに従ってください。

## 参考リンク

- [AquesTalk公式ブログ](http://blog-yama.a-quest.com/?eid=970181)
- [AquesTalk FAQ](https://www.a-quest.com/faq.html)