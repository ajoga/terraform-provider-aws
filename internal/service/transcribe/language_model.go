package transcribe

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/transcribe"
	"github.com/aws/aws-sdk-go-v2/service/transcribe/types"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceLanguageModel() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLanguageModelCreate,
		ReadWithoutTimeout:   resourceLanguageModelRead,
		UpdateWithoutTimeout: resourceLanguageModelUpdate,
		DeleteWithoutTimeout: resourceLanguageModelDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"base_model_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(baseModelNamesSlice(types.BaseModelName("").Values()), false),
			},
			"input_data_config": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_access_role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
						"s3_uri": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"tuning_data_s3_uri": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Computed: true,
						},
					},
				},
			},
			"language_code": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(languageCodeSlice(types.LanguageCode("").Values()), false),
			},
			"model_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceLanguageModelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).TranscribeConn

	in := &transcribe.CreateLanguageModelInput{
		BaseModelName: types.BaseModelName(d.Get("base_model_name").(string)),
		LanguageCode:  types.CLMLanguageCode(d.Get("language_code").(string)),
		ModelName:     aws.String(d.Get("model_name").(string)),
	}

	if v, ok := d.GetOk("input_data_config"); ok && len(v.([]interface{})) > 0 {
		in.InputDataConfig = expandInputDataConfig(v.([]interface{}))
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateLanguageModel(ctx, in)
	if err != nil {
		return diag.Errorf("creating Amazon Transcribe LanguageModel (%s): %s", d.Get("model_name").(string), err)
	}

	d.SetId(aws.ToString(out.ModelName))

	if _, err := waitLanguageModelCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Amazon Transcribe LanguageModel (%s) create: %s", d.Id(), err)
	}

	return resourceLanguageModelRead(ctx, d, meta)
}

func resourceLanguageModelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).TranscribeConn

	out, err := FindLanguageModelByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Transcribe LanguageModel (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Transcribe LanguageModel (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "transcribe",
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("/language-model/%s", d.Id()),
	}.String()

	d.Set("arn", arn)
	d.Set("base_model_name", out.BaseModelName)
	d.Set("language_code", out.LanguageCode)
	d.Set("model_name", out.ModelName)

	if err := d.Set("input_data_config", flattenInputDataConfig(out.InputDataConfig)); err != nil {
		return diag.Errorf("setting input data config: %s", err)
	}

	tags, err := ListTags(ctx, conn, d.Id())
	if err != nil {
		return diag.Errorf("listing tags for Transcribe LanguageModel (%s): %s", d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceLanguageModelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).TranscribeConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("error updating Transcribe LanguageModel (%s) tags: %s", d.Get("db_snapshot_arn").(string), err)
		}
	}

	return resourceLanguageModelRead(ctx, d, meta)
}

func resourceLanguageModelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).TranscribeConn

	log.Printf("[INFO] Deleting Transcribe LanguageModel %s", d.Id())

	_, err := conn.DeleteLanguageModel(ctx, &transcribe.DeleteLanguageModelInput{
		ModelName: aws.String(d.Id()),
	})

	var resourceNotFoundException *types.NotFoundException
	if errors.As(err, &resourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Transcribe LanguageModel (%s): %s", d.Id(), err)
	}

	return nil
}

func waitLanguageModelCreated(ctx context.Context, conn *transcribe.Client, id string, timeout time.Duration) (*types.LanguageModel, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   modelStatus(types.ModelStatusInProgress),
		Target:                    modelStatus(types.ModelStatusCompleted),
		Refresh:                   statusLanguageModel(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.LanguageModel); ok {
		return out, err
	}

	return nil, err
}

func statusLanguageModel(ctx context.Context, conn *transcribe.Client, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindLanguageModelByName(ctx, conn, name)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.ModelStatus), nil
	}
}

func FindLanguageModelByName(ctx context.Context, conn *transcribe.Client, id string) (*types.LanguageModel, error) {
	in := &transcribe.DescribeLanguageModelInput{
		ModelName: aws.String(id),
	}

	out, err := conn.DescribeLanguageModel(ctx, in)

	var resourceNotFoundException *types.NotFoundException
	if errors.As(err, &resourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.LanguageModel == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.LanguageModel, nil
}

func flattenInputDataConfig(apiObjects *types.InputDataConfig) []interface{} {
	if apiObjects == nil {
		return nil
	}

	m := map[string]interface{}{
		"data_access_role_arn": apiObjects.DataAccessRoleArn,
		"s3_uri":               apiObjects.S3Uri,
		"tuning_data_s3_uri":   apiObjects.TuningDataS3Uri,
	}

	return []interface{}{m}
}

func expandInputDataConfig(tfList []interface{}) *types.InputDataConfig {
	var s *types.InputDataConfig

	i := tfList[0].(map[string]interface{})

	if val, ok := i["data_access_role_arn"]; ok {
		s.DataAccessRoleArn = aws.String(val.(string))
	}

	if val, ok := i["s3_uri"]; ok {
		s.S3Uri = aws.String(val.(string))
	}

	if val, ok := i["tuning_data_s3_uri"]; ok {
		s.TuningDataS3Uri = aws.String(val.(string))
	}

	return s
}

func modelStatus(in ...types.ModelStatus) []string {
	var s []string

	for _, v := range in {
		s = append(s, string(v))
	}

	return s
}

func baseModelNamesSlice(in []types.BaseModelName) (out []string) {
	for _, v := range in {
		out = append(out, string(v))
	}

	return
}

func languageCodeSlice(in []types.LanguageCode) (out []string) {
	for _, v := range in {
		out = append(out, string(v))
	}

	return
}
